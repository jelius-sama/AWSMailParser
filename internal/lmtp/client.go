package lmtp

import (
    "AWSMailParser/internal/addons"
    "AWSMailParser/vars"
    "bufio"
    "fmt"
    "net"
    "os"
    "strconv"
    "strings"

    "github.com/jelius-sama/logger"
)

func SendLMTPCommand(writer *bufio.Writer, reader *bufio.Reader, command string) (string, error) {
    if _, err := writer.WriteString(command + "\r\n"); err != nil {
        return "", fmt.Errorf("failed to send command: %w", err)
    }
    if err := writer.Flush(); err != nil {
        return "", fmt.Errorf("failed to flush command: %w", err)
    }

    response, err := reader.ReadString('\n')
    if err != nil {
        return "", fmt.Errorf("failed to read response: %w", err)
    }

    return strings.TrimSpace(response), nil
}

func ReadLMTPMultilineResponse(reader *bufio.Reader) error {
    for {
        line, err := reader.ReadString('\n')
        if err != nil {
            return fmt.Errorf("failed to read response: %w", err)
        }

        // Check if this is the last line (code without '-')
        if len(line) >= 4 && line[3] != '-' {
            if code, _ := strconv.Atoi(line[:3]); code >= 400 {
                return fmt.Errorf("command failed: %s", strings.TrimSpace(line))
            }
            break
        }
    }
    return nil
}

func DeliverViaLMTP(fromAddr string, rcptList []string, rawBytes []byte) error {
    if len(rcptList) == 0 {
        return fmt.Errorf("no recipients provided")
    }
    if len(rawBytes) > vars.MaxMessageSize {
        return fmt.Errorf("message too large: %d bytes", len(rawBytes))
    }

    conn, err := net.DialTimeout("tcp", net.JoinHostPort(vars.LMTPHost, strconv.Itoa(vars.LMTPPort)), vars.LMTPTimeout)
    if err != nil {
        logger.TimedError("LMTP connection failed:", err)
        return err
    }
    defer conn.Close()

    reader := bufio.NewReader(conn)
    writer := bufio.NewWriter(conn)

    // Read greeting
    if _, err := reader.ReadString('\n'); err != nil {
        return fmt.Errorf("failed to read greeting: %w", err)
    }

    // Send LHLO command
    hostname, _ := os.Hostname()
    if hostname == "" {
        hostname = "localhost"
    }

    if _, err := writer.WriteString(fmt.Sprintf("LHLO %s\r\n", hostname)); err != nil {
        return fmt.Errorf("failed to send LHLO: %w", err)
    }
    if err := writer.Flush(); err != nil {
        return fmt.Errorf("failed to flush LHLO: %w", err)
    }

    // Read LHLO multiline response
    if err := ReadLMTPMultilineResponse(reader); err != nil {
        return fmt.Errorf("LHLO failed: %w", err)
    }

    // Send MAIL FROM
    response, err := SendLMTPCommand(writer, reader, fmt.Sprintf("MAIL FROM:<%s>", fromAddr))
    if err != nil {
        return fmt.Errorf("MAIL FROM failed: %w", err)
    }
    if code, _ := strconv.Atoi(response[:3]); code >= 400 {
        return fmt.Errorf("MAIL FROM failed: %s", response)
    }

    // Send RCPT TO for each recipient
    var failed []string
    for _, rcpt := range rcptList {
        response, err := SendLMTPCommand(writer, reader, fmt.Sprintf("RCPT TO:<%s>", rcpt))
        if err != nil || (len(response) >= 3 && func() bool {
            code, _ := strconv.Atoi(response[:3])
            return code >= 400
        }()) {
            logger.TimedError("RCPT TO failed for", rcpt, ":", response)
            failed = append(failed, rcpt)
        }
    }

    if len(failed) > 0 {
        logger.TimedError("LMTP delivery failed for", failed)
    }

    if len(failed) == len(rcptList) {
        return fmt.Errorf("all recipients failed")
    }

    // Send DATA command
    response, err = SendLMTPCommand(writer, reader, "DATA")
    if err != nil {
        return fmt.Errorf("DATA failed: %w", err)
    }
    if code, _ := strconv.Atoi(response[:3]); code >= 400 {
        return fmt.Errorf("DATA failed: %s", response)
    }

    // Send message data
    if _, err := writer.Write(rawBytes); err != nil {
        return fmt.Errorf("failed to write message: %w", err)
    }

    // Send end-of-data marker
    response, err = SendLMTPCommand(writer, reader, ".")
    if err != nil {
        return fmt.Errorf("message delivery failed: %w", err)
    }
    if code, _ := strconv.Atoi(response[:3]); code >= 400 {
        return fmt.Errorf("message delivery failed: %s", response)
    }

    // Send QUIT (ignore errors)
    writer.WriteString("QUIT\r\n")
    writer.Flush()

    if len(failed) == 0 {
        logger.Okay("LMTP delivery succeeded for", rcptList)
        if addons.IsFromBank(fromAddr) {
            logger.Info("Detected mail from bank, sending it to Zaimu.")
            if err := addons.ForwardToZaimu(rawBytes, fromAddr); err != nil {
                logger.Error("Forwarding may have failed.\nZaimu returned:\n\t", err)
            }
        }
    }

    return nil
}

