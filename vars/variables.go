package vars

import (
    "path/filepath"
    "regexp"

    "github.com/aws/aws-sdk-go/aws/session"
    "github.com/aws/aws-sdk-go/service/s3"
    "github.com/aws/aws-sdk-go/service/sqs"
)

var (
    SESS      *session.Session
    S3Client  *s3.S3
    SQSClient *sqs.SQS

    // Regex for parsing Received headers
    ReceivedForRegex *regexp.Regexp

    // Unix Socket Path for Zaimu addon
    ZaimuUnixSockPath = filepath.Join("/home/kazuma", "zaimu", "unix.sock")
)

