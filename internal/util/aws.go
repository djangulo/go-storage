// Package awsabs has dependency injected abstractions for AWS-S3 services.
package util

import (
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3iface"
)

func BucketExists(client s3iface.S3API, bucket string) (exists bool) {
	if bucket == "" {
		return false
	}
	_, err := client.HeadBucket(&s3.HeadBucketInput{
		Bucket: &bucket,
	})
	if aerr, ok := err.(awserr.Error); ok {
		switch aerr.Code() {
		case "NotFound":
		case s3.ErrCodeNoSuchBucket:
			exists = false
		default:
			exists = true
		}
	}
	if err == nil {
		exists = true
	}
	return
}
