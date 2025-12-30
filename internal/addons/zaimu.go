package addons

import (
    "AWSMailParser/vars"
    "bytes"
    "context"
    "errors"
    "fmt"
    "net"
    "net/http"
    "strings"
)

func IsFromBank(sender string) bool {
    var knownTransactionSenders = []string{
        "no-reply@google.com",
        "alerts@hdfcbank.net",
        "alerts@icicibank.com",
        "noreply@axisbank.com",
        "alerts@chase.com",
        "alerts@bankofamerica.com",
    }

    for _, s := range knownTransactionSenders {
        if strings.EqualFold(sender, s) {
            return true
        }
    }
    return false
}

func ForwardToZaimu(
    rawEML []byte,
    sender string,
) error {
    transport := &http.Transport{
        DialContext: func(_ context.Context, _, _ string) (net.Conn, error) {
            return net.Dial("unix", vars.ZaimuUnixSockPath)
        },
    }

    client := &http.Client{
        Transport: transport,
    }

    req, err := http.NewRequest(
        http.MethodPost,
        "http://unix/ingest/email",
        bytes.NewReader(rawEML),
    )
    if err != nil {
        return err
    }

    req.Header.Set("Content-Type", "application/octet-stream")

    resp, err := client.Do(req)
    if err != nil {
        return err
    }
    defer resp.Body.Close()

    if resp.StatusCode >= 300 {
        return errors.New(fmt.Sprintln(resp))
    }

    return nil
}

