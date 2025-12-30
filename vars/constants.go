package vars

import (
    "time"
)

const (
    Region      = "ap-northeast-1"
    AccountID   = "000000000000"
    S3Bucket    = "your-mail-bucket"
    SQSEvent    = "your-mail-events"
    MaildirBase = "/var/vmail"

    SQSQueueURL = "https://sqs." + Region + ".amazonaws.com/" + AccountID + "/" + SQSEvent

    LMTPHost       = "127.0.0.1"
    LMTPPort       = 2003
    MaxRetries     = 3
    MaxMessageSize = 40 * 1024 * 1024 // 40MB

    // LMTP timeout
    LMTPTimeout = 60 * time.Second

    // SQS polling configuration
    MaxMessages          = 5
    WaitTimeSeconds      = 20
    VisibilityTimeout    = 300
    MaxConsecutiveErrors = 5
)

