package testcoverage_test

import (
	"net/http/httptest"
	"os"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/johannesboyne/gofakes3"
	"github.com/johannesboyne/gofakes3/backend/s3mem"
	"github.com/stretchr/testify/assert"

	. "github.com/vladopajic/go-test-coverage/v2/pkg/testcoverage"
)

func Test_GenerateAndSaveBadge_NoAction(t *testing.T) {
	t.Parallel()

	// should not return error when badge file name is not specified
	err := GenerateAndSaveBadge(Config{
		Badge: Badge{},
	}, 100)
	assert.NoError(t, err)
}

func Test_GenerateAndSaveBadge_SaveToFile(t *testing.T) {
	t.Parallel()

	if testing.Short() {
		return
	}

	// should save badge to file
	testFile := t.TempDir() + "/badge.svg"

	err := GenerateAndSaveBadge(Config{
		Badge: Badge{
			FileName: testFile,
		},
	}, 100)
	assert.NoError(t, err)

	contentBytes, err := os.ReadFile(testFile)
	assert.NoError(t, err)
	assert.NotEmpty(t, contentBytes)
}

func Test_GenerateAndSaveBadge_SaveToCDN(t *testing.T) {
	t.Parallel()

	if testing.Short() {
		return
	}

	const (
		key    = `🔑`
		secret = `your-secrets-are-safu`
	)

	// key not prvided
	err := GenerateAndSaveBadge(Config{
		Badge: Badge{
			CDN: CDN{Secret: secret},
		},
	}, 100)
	assert.Error(t, err)

	backend := s3mem.New()
	faker := gofakes3.New(backend)
	ts := httptest.NewServer(faker.Server())

	defer ts.Close()

	cdn := CDN{
		Key:            key,
		Secret:         secret,
		Region:         "eu-central-1",
		FileName:       "coverage.svg",
		BucketName:     "badges",
		Endpoint:       ts.URL,
		ForcePathStyle: true,
	}

	// bucket does not exists
	err = GenerateAndSaveBadge(Config{Badge: Badge{CDN: cdn}}, 100)
	assert.Error(t, err)

	// create bucket and try again
	s3Client, err := CreateS3Client(cdn)
	assert.NoError(t, err)

	_, err = s3Client.CreateBucket(&s3.CreateBucketInput{
		Bucket: aws.String(cdn.BucketName),
	})
	assert.NoError(t, err)

	// bucket exists
	err = GenerateAndSaveBadge(Config{Badge: Badge{CDN: cdn}}, 100)
	assert.NoError(t, err)
}
