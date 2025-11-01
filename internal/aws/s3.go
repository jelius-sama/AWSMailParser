package aws

import (
	"github.com/aws/aws-sdk-go/service/s3"

	"AWSMailParser/vars"
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/jelius-sama/logger"
	"io"
	"net/url"
	"strings"
	"time"
)

func (m *Mailfetcher) ExtractS3Info(eventJSON map[string]any) (bucket, key string, err error) {
	// Handle S3 test events
	if event, ok := eventJSON["Event"].(string); ok && event == "s3:TestEvent" {
		logger.TimedInfo("Skipping S3 test event")
		return "", "", fmt.Errorf("test event")
	}

	records, ok := eventJSON["Records"].([]any)
	if !ok {
		logger.TimedError("No 'Records' field in event JSON")
		return "", "", fmt.Errorf("no records found")
	}

	if len(records) == 0 {
		return "", "", fmt.Errorf("empty records")
	}

	record, ok := records[0].(map[string]any)
	if !ok {
		return "", "", fmt.Errorf("invalid record format")
	}

	s3Info, ok := record["s3"].(map[string]any)
	if !ok {
		logger.TimedError("No 's3' field in record")
		return "", "", fmt.Errorf("no s3 field")
	}

	bucketInfo, ok := s3Info["bucket"].(map[string]any)
	if !ok {
		return "", "", fmt.Errorf("no bucket info")
	}

	bucket, ok = bucketInfo["name"].(string)
	if !ok {
		return "", "", fmt.Errorf("no bucket name")
	}

	objectInfo, ok := s3Info["object"].(map[string]any)
	if !ok {
		return "", "", fmt.Errorf("no object info")
	}

	key, ok = objectInfo["key"].(string)
	if !ok {
		return "", "", fmt.Errorf("no object key")
	}

	// URL decode the key (equivalent to Python's unquote_plus)
	decodedKey, err := url.QueryUnescape(strings.ReplaceAll(key, "+", " "))
	if err != nil {
		logger.TimedError("Failed to decode key:", err)
		return "", "", err
	}

	return bucket, decodedKey, nil
}

func (m *Mailfetcher) FetchFromS3(bucket, key string) ([]byte, error) {
	for attempt := range vars.MaxRetries {
		obj, err := vars.S3Client.GetObject(&s3.GetObjectInput{
			Bucket: aws.String(bucket),
			Key:    aws.String(key),
		})
		if err != nil {
			logger.TimedError("S3 fetch attempt", attempt+1, " failed:", err)
			if attempt == vars.MaxRetries-1 {
				return nil, err
			}
			time.Sleep(time.Duration(1<<attempt) * time.Second)
			continue
		}

		defer obj.Body.Close()
		return io.ReadAll(obj.Body)
	}

	return nil, fmt.Errorf("all S3 fetch attempts failed")
}
