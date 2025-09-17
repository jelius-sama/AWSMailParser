package email

import (
	"AWSMailParser/logger"
	"AWSMailParser/vars"
	"github.com/DusanKasan/parsemail"
)

func ExtractRecipientsFromHeaders(mail *parsemail.Email) []string {
	// Try Delivered-To header first
	if deliveredTo := mail.Header.Get("Delivered-To"); deliveredTo != "" {
		logger.TimedInfo("Using Delivered-To recipient:", deliveredTo)
		return []string{deliveredTo}
	}

	// Try X-Original-To
	if xOriginalTo := mail.Header.Get("X-Original-To"); xOriginalTo != "" {
		logger.TimedInfo("Using X-Original-To recipient:", xOriginalTo)
		return []string{xOriginalTo}
	}

	// Parse Received headers for "for <...>"
	for _, line := range mail.Header["Received"] {
		if matches := vars.ReceivedForRegex.FindStringSubmatch(line); len(matches) > 1 {
			logger.TimedInfo("Using recipient from Received header:", matches[1])
			return []string{matches[1]}
		}
	}

	// Fallback to To/Cc/Bcc
	var rcpts []string
	allAddresses := append(append(mail.To, mail.Cc...), mail.Bcc...)
	for _, addr := range allAddresses {
		if addr.Address != "" {
			rcpts = append(rcpts, addr.Address)
		}
	}

	logger.TimedInfo("Using header recipients:", rcpts)
	return rcpts
}

func ExtractRecipients(eventJSON map[string]any, mail *parsemail.Email) []string {
	// Try SES recipients first
	if records, ok := eventJSON["Records"].([]any); ok && len(records) > 0 {
		if record, ok := records[0].(map[string]any); ok {
			// Check multiple paths for recipients
			for _, path := range [][]string{
				{"ses", "receipt", "recipients"},
				{"receipt", "recipients"},
			} {
				if rcpts := ExtractRecipientsFromPath(record, path); len(rcpts) > 0 {
					logger.TimedInfo("Using SES recipients:", rcpts)
					return rcpts
				}
			}
		}
	}

	return ExtractRecipientsFromHeaders(mail)
}

func ExtractRecipientsFromPath(data map[string]any, path []string) []string {
	current := data
	for _, key := range path[:len(path)-1] {
		next, ok := current[key].(map[string]any)
		if !ok {
			return nil
		}
		current = next
	}

	recipients, ok := current[path[len(path)-1]].([]any)
	if !ok || len(recipients) == 0 {
		return nil
	}

	var result []string
	for _, r := range recipients {
		if rcpt, ok := r.(string); ok {
			result = append(result, rcpt)
		}
	}

	return result
}
