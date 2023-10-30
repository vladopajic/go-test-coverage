package testcoverage

import (
	"bytes"
	"fmt"
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"

	"github.com/vladopajic/go-test-coverage/v2/pkg/testcoverage/badge"
)

func GenerateAndSaveBadge(cfg Config, totalCoverage int) error {
	badge, err := badge.Generate(totalCoverage)
	if err != nil {
		return fmt.Errorf("generate badge: %w", err)
	}

	if cfg.Badge.FileName != "" {
		err := saveBadeToFile(cfg.Badge.FileName, badge)
		if err != nil {
			return fmt.Errorf("save badge to file: %w", err)
		}
	}

	if cfg.Badge.CDN.Secret != "" {
		err := saveBadgeToCDN(cfg.Badge.CDN, badge)
		if err != nil {
			return fmt.Errorf("save badge to cdn: %w", err)
		}
	}

	return nil
}

func saveBadeToFile(filename string, data []byte) error {
	return os.WriteFile(filename, data, 0o644) //nolint:gosec,gomnd,wrapcheck // relax
}

type CDN struct {
	Key            string
	Secret         string
	Region         string
	FileName       string
	BucketName     string
	Endpoint       string
	ForcePathStyle bool
}

func saveBadgeToCDN(cdn CDN, data []byte) error {
	s3Client, err := createS3Client(cdn)
	if err != nil {
		return fmt.Errorf("create s3 client: %w", err)
	}

	object := s3.PutObjectInput{
		Bucket:        aws.String(cdn.BucketName),
		Key:           aws.String(cdn.FileName),
		Body:          bytes.NewReader(data),
		ContentType:   aws.String(badge.ContentType),
		ContentLength: aws.Int64(int64(len(data))),
	}

	_, err = s3Client.PutObject(&object)
	if err != nil {
		return fmt.Errorf("put object: %w", err)
	}

	return nil
}

func createS3Client(cdn CDN) (*s3.S3, error) {
	s3Config := &aws.Config{
		Credentials:      credentials.NewStaticCredentials(cdn.Key, cdn.Secret, ""),
		Endpoint:         aws.String(cdn.Endpoint),
		Region:           aws.String(cdn.Region),
		S3ForcePathStyle: aws.Bool(cdn.ForcePathStyle),
	}

	newSession, err := session.NewSession(s3Config)
	if err != nil {
		return nil, fmt.Errorf("create session: %w", err)
	}

	s3Client := s3.New(newSession)

	return s3Client, nil
}
