package main

import (
	"AWSMailParser/internal/aws"
	"AWSMailParser/vars"
	"context"
	"github.com/jelius-sama/logger"
	"os/signal"
	"syscall"

	awsSDK "github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/sqs"
	"regexp"
)

func init() {
	vars.SESS = session.Must(session.NewSession(&awsSDK.Config{
		Region: awsSDK.String(vars.Region),
	}))

	vars.S3Client = s3.New(vars.SESS)
	vars.SQSClient = sqs.New(vars.SESS)
	vars.ReceivedForRegex = regexp.MustCompile(`for <([^>]+)>`)
}

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(),
		syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	fetcher := &aws.Mailfetcher{}

	if err := fetcher.PollSQS(ctx); err != nil && err != context.Canceled {
		logger.TimedPanic("Fatal error:", err)
	}
}
