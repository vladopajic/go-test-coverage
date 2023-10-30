package testcoverage

import (
	"bytes"
	"fmt"
	"os"
	"strings"

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
	Key        string
	Secret     string
	Region     string
	FileName   string
	BucketName string
	Endpoint   string
}

func saveBadgeToCDN(cdn CDN, data []byte) error {
	if strings.Contains(cdn.Endpoint, ".digitaloceanspaces.com") {
		cdn.Region = "us-east-1" // for spaces region must be us-east-1
	}

	s3Config := &aws.Config{
		Credentials:      credentials.NewStaticCredentials(cdn.Key, cdn.Secret, ""),
		Endpoint:         aws.String(cdn.Endpoint),
		Region:           aws.String(cdn.Region),
		S3ForcePathStyle: aws.Bool(false),
	}

	newSession, err := session.NewSession(s3Config)
	if err != nil {
		return fmt.Errorf("create session: %w", err)
	}

	s3Client := s3.New(newSession)

	object := s3.PutObjectInput{
		Bucket: aws.String(cdn.BucketName),
		Key:    aws.String(cdn.FileName),
		Body:   bytes.NewReader(data),
	}

	_, err = s3Client.PutObject(&object)
	if err != nil {
		return fmt.Errorf("put object: %w", err)
	}

	return nil
}
