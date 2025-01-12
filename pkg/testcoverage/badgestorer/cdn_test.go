package badgestorer_test

import (
	"net/http/httptest"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/johannesboyne/gofakes3"
	"github.com/johannesboyne/gofakes3/backend/s3mem"
	"github.com/stretchr/testify/assert"

	. "github.com/subhambhardwaj/go-test-coverage/v2/pkg/testcoverage/badgestorer"
)

func Test_CDN_Error(t *testing.T) {
	t.Parallel()

	data := []byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 0}
	cfg := CDN{
		Secret: `your-secrets-are-safu`,
	}

	s := NewCDN(cfg)
	updated, err := s.Store(data)
	assert.Error(t, err)
	assert.False(t, updated)
}

func Test_CDN(t *testing.T) {
	t.Parallel()

	if testing.Short() {
		return
	}

	data := []byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 0}

	const (
		key      = `🔑`
		secret   = `your-secrets-are-safu`
		coverage = 100
	)

	backend := s3mem.New()
	faker := gofakes3.New(backend)
	ts := httptest.NewServer(faker.Server())

	defer ts.Close()

	cfg := CDN{
		Key:            key,
		Secret:         secret,
		Region:         "eu-central-1",
		FileName:       "coverage.svg",
		BucketName:     "badges",
		Endpoint:       ts.URL,
		ForcePathStyle: true,
	}

	// bucket does not exists
	s := NewCDN(cfg)
	updated, err := s.Store(data)
	assert.Error(t, err)
	assert.False(t, updated)

	// create bucket and assert again
	s3Client := CreateS3Client(cfg)

	_, err = s3Client.CreateBucket(&s3.CreateBucketInput{
		Bucket: aws.String(cfg.BucketName),
	})
	assert.NoError(t, err)

	// put badge
	updated, err = s.Store(data)
	assert.NoError(t, err)
	assert.True(t, updated)

	// put badge again - no change
	updated, err = s.Store(data)
	assert.NoError(t, err)
	assert.False(t, updated)

	// put badge again - expect change
	updated, err = s.Store(append(data, byte(1)))
	assert.NoError(t, err)
	assert.True(t, updated)
}
