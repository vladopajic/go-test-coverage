package badgestorer

import (
	"bytes"
	"fmt"
	"io"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"

	"github.com/subhambhardwaj/go-test-coverage/v2/pkg/testcoverage/badge"
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

	// First get object and check if data differs that currently uploaded
	result, err := s3Client.GetObject(&s3.GetObjectInput{
		Bucket: aws.String(s.cfg.BucketName),
		Key:    aws.String(s.cfg.FileName),
	})
	if err == nil {
		//nolint:errcheck // error is intentionally swallowed because if response (badge data)
		// is not the same we will upload new badge anyway
		resp, _ := io.ReadAll(result.Body)
		if bytes.Equal(resp, data) {
			return false, nil // has not changed
		}
	}

	// Currently uploaded badge does not exists or has changed
	// so it should be uploaded
	_, err = s3Client.PutObject(&s3.PutObjectInput{
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

func createS3Client(cfg CDN) *s3.S3 {
	s3Config := &aws.Config{
		Credentials:      credentials.NewStaticCredentials(cfg.Key, cfg.Secret, ""),
		Endpoint:         aws.String(cfg.Endpoint),
		Region:           aws.String(cfg.Region),
		S3ForcePathStyle: aws.Bool(cfg.ForcePathStyle),
	}

	// calling `session.Must` can potentially panic, which is not practice of this
	// codebase to panic outside of main function. however it will never happen as
	// this panic only happens when sessions could not be created using env variables.
	newSession := session.Must(session.NewSession(s3Config))

	return s3.New(newSession)
}
