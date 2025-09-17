package aws

import (
	"AWSMailParser/internal/email"
	"AWSMailParser/internal/lmtp"
	"AWSMailParser/internal/maildir"
	"AWSMailParser/logger"
	"AWSMailParser/vars"
	"bytes"
	"fmt"
	"strings"

	"github.com/DusanKasan/parsemail"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/sqs"
)

func (m *Mailfetcher) DeleteMessage(receiptHandle *string) {
	if _, err := vars.SQSClient.DeleteMessage(&sqs.DeleteMessageInput{
		QueueUrl:      aws.String(vars.SQSQueueURL),
		ReceiptHandle: receiptHandle,
	}); err != nil {
		logger.TimedError("Failed to delete message:", err)
	} else {
		logger.TimedOkay("Message processed and deleted")
	}
}

func (m *Mailfetcher) HandleMessages(messages []*sqs.Message) {
	m.ConsecutiveErrors = 0

	for _, msg := range messages {
		if m.ProcessMessage(msg) {
			m.DeleteMessage(msg.ReceiptHandle)
		} else {
			logger.TimedError("Message processing failed, will retry later")
		}
	}
}

func (m *Mailfetcher) ProcessMessage(msg *sqs.Message) bool {
	body := *msg.Body
	logger.TimedInfo("Processing SQS message:", body[:min(len(body), 200)])

	eventJSON, err := m.ParseSQSBody(body)
	if err != nil {
		return false
	}

	bucket, key, err := m.ExtractS3Info(eventJSON)
	if err != nil {
		return false
	}

	rawBytes, err := m.FetchFromS3(bucket, key)
	if err != nil {
		return false
	}

	mail, err := parsemail.Parse(bytes.NewReader(rawBytes))
	if err != nil {
		logger.TimedError("Failed to parse email:", err)
		return false
	}

	recipients := email.ExtractRecipients(eventJSON, &mail)
	if len(recipients) == 0 {
		logger.TimedError("No valid recipients found for key:", key)
		return false
	}

	// Create maildirs for all recipients
	for _, recipient := range recipients {
		if parts := strings.SplitN(recipient, "@", 2); len(parts) == 2 {
			local, domain := parts[0], parts[1]
			if err := maildir.EnsureMaildir(domain, local); err != nil {
				logger.TimedError("Failed to create maildir for", recipient, ":", err)
			}
		}
	}

	// Determine from address
	fromAddr := ""
	if len(mail.From) > 0 {
		fromAddr = mail.From[0].Address
	}
	if fromAddr == "" && len(recipients) > 0 {
		if parts := strings.SplitN(recipients[0], "@", 2); len(parts) == 2 {
			fromAddr = fmt.Sprintf("postmaster@%s", parts[1])
		}
	}

	logger.TimedInfo("Using envelope from:", fromAddr)

	if err := lmtp.DeliverViaLMTP(fromAddr, recipients, rawBytes); err != nil {
		logger.TimedError("Failed to deliver email via LMTP:", err)
		return false
	}

	return true
}
