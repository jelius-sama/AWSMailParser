package aws

import (
	"AWSMailParser/vars"
	"context"
	"encoding/json"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/jelius-sama/logger"
	"time"

	"github.com/aws/aws-sdk-go/service/sqs"
)

func (m *Mailfetcher) HandleError(err error) bool {
	m.ConsecutiveErrors++
	logger.TimedError("SQS polling error (attempt", m.ConsecutiveErrors, "):", err)

	if m.ConsecutiveErrors >= vars.MaxConsecutiveErrors {
		logger.TimedPanic("Too many consecutive errors. Exiting.")
		return false
	}

	sleepTime := min(60, 1<<m.ConsecutiveErrors)
	time.Sleep(time.Duration(sleepTime) * time.Second)
	return true
}

func (m *Mailfetcher) PollSQS(ctx context.Context) error {
	logger.Info("mailfetcher started, polling SQS:", vars.SQSQueueURL)

	for {
		select {
		case <-ctx.Done():
			logger.TimedInfo("Shutting down mailfetcher...")
			return ctx.Err()
		default:
		}

		resp, err := vars.SQSClient.ReceiveMessage(&sqs.ReceiveMessageInput{
			QueueUrl:              aws.String(vars.SQSQueueURL),
			MaxNumberOfMessages:   aws.Int64(vars.MaxMessages),
			WaitTimeSeconds:       aws.Int64(vars.WaitTimeSeconds),
			VisibilityTimeout:     aws.Int64(vars.VisibilityTimeout),
			MessageAttributeNames: aws.StringSlice([]string{"All"}),
		})

		if err != nil {
			if !m.HandleError(err) {
				return err
			}
			continue
		}

		if messages := resp.Messages; len(messages) > 0 {
			m.HandleMessages(messages)
		}
	}
}

func (m *Mailfetcher) ParseSQSBody(body string) (map[string]any, error) {
	var msg map[string]any
	if err := json.Unmarshal([]byte(body), &msg); err != nil {
		logger.TimedError("Failed to parse JSON body:", err)
		return nil, err
	}

	// Handle SNS message wrapping
	if message, ok := msg["Message"].(string); ok {
		var innerMsg map[string]any
		if err := json.Unmarshal([]byte(message), &innerMsg); err != nil {
			logger.TimedError("Failed to parse SNS Message:", err)
			return nil, err
		}
		return innerMsg, nil
	}
	return msg, nil
}
