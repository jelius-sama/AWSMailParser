package vars

import (
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/sqs"
	"regexp"
)

var (
	SESS      *session.Session
	S3Client  *s3.S3
	SQSClient *sqs.SQS

	// Regex for parsing Received headers
	ReceivedForRegex *regexp.Regexp
)
