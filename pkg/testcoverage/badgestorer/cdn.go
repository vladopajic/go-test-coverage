package badgestorer

import (
	"bytes"
	"context"
	"fmt"
	"io"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"

	"github.com/vladopajic/go-test-coverage/v2/pkg/testcoverage/badge"
)

type CDN struct {
	Key            string
	Secret         string
	Region         string
	FileName       string
	BucketName     string
	Endpoint       string
	ForcePathStyle bool
}

type cdnStorer struct {
	cfg CDN
}

func NewCDN(cfg CDN) Storer {
	return &cdnStorer{cfg: cfg}
}

func (s *cdnStorer) Store(data []byte) (bool, error) {
	s3Client := createS3Client(s.cfg)
	ctx := context.Background()

	// First get object and check if data differs that currently uploaded
	result, err := s3Client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(s.cfg.BucketName),
		Key:    aws.String(s.cfg.FileName),
	})
	if err == nil {
		defer result.Body.Close()

		//nolint:errcheck // error is intentionally swallowed because if response (badge data)
		// is not the same we will upload new badge anyway
		resp, _ := io.ReadAll(result.Body)
		if bytes.Equal(resp, data) {
			return false, nil // has not changed
		}
	}

	// Currently uploaded badge does not exists or has changed
	// so it should be uploaded
	_, err = s3Client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:        aws.String(s.cfg.BucketName),
		Key:           aws.String(s.cfg.FileName),
		Body:          bytes.NewReader(data),
		ContentType:   aws.String(badge.ContentType),
		ContentLength: aws.Int64(int64(len(data))),
	})
	if err != nil {
		return false, fmt.Errorf("put object: %w", err)
	}

	return true, nil // has changed
}

func createS3Client(cfg CDN) *s3.Client {
	credentialsProvider := credentials.NewStaticCredentialsProvider(cfg.Key, cfg.Secret, "")
	sdkConfig := aws.Config{
		Credentials: aws.NewCredentialsCache(credentialsProvider),
		Region:      cfg.Region,
	}

	return s3.NewFromConfig(sdkConfig, func(options *s3.Options) {
		options.BaseEndpoint = aws.String(cfg.Endpoint)
		options.UsePathStyle = cfg.ForcePathStyle
	})
}
