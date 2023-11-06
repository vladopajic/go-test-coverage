package testcoverage_test

import (
	"bytes"
	"io"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/johannesboyne/gofakes3"
	"github.com/johannesboyne/gofakes3/backend/s3mem"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	. "github.com/vladopajic/go-test-coverage/v2/pkg/testcoverage"
	"github.com/vladopajic/go-test-coverage/v2/pkg/testcoverage/badge"
)

func Test_GenerateAndSaveBadge_NoAction(t *testing.T) {
	t.Parallel()

	// should not return error when badge file name is not specified
	err := GenerateAndSaveBadge(nil, Config{
		Badge: Badge{},
	}, 100)
	assert.NoError(t, err)
}

func Test_GenerateAndSaveBadge_SaveToFile(t *testing.T) {
	t.Parallel()

	if testing.Short() {
		return
	}

	t.Run("invalid file", func(t *testing.T) {
		err := GenerateAndSaveBadge(nil, Config{
			Badge: Badge{
				FileName: t.TempDir(), // should not be able to write to directory
			},
		}, 100)
		assert.Error(t, err)
	})

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		// should save badge to file
		testFile := t.TempDir() + "/badge.svg"

		buf := &bytes.Buffer{}
		err := GenerateAndSaveBadge(buf, Config{
			Badge: Badge{
				FileName: testFile,
			},
		}, 100)
		assert.NoError(t, err)
		assert.Contains(t, buf.String(), "Badge saved to file")

		contentBytes, err := os.ReadFile(testFile)
		assert.NoError(t, err)
		assert.NotEmpty(t, contentBytes)
	})
}

func Test_GenerateAndSaveBadge_SaveToCDN_NoAction(t *testing.T) {
	t.Parallel()

	// key not prvided
	err := GenerateAndSaveBadge(nil,
		Config{
			Badge: Badge{
				CDN: CDN{Secret: `your-secrets-are-safu`},
			},
		}, 100)
	assert.Error(t, err)
}

func Test_GenerateAndSaveBadge_SaveToCDN(t *testing.T) {
	t.Parallel()

	if testing.Short() {
		return
	}

	const (
		key      = `🔑`
		secret   = `your-secrets-are-safu`
		coverage = 100
	)

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
	err := GenerateAndSaveBadge(nil, Config{Badge: Badge{CDN: cdn}}, coverage)
	assert.Error(t, err)

	// create bucket and assert again
	s3Client, err := CreateS3Client(cdn)
	require.NoError(t, err)

	_, err = s3Client.CreateBucket(&s3.CreateBucketInput{
		Bucket: aws.String(cdn.BucketName),
	})
	assert.NoError(t, err)

	// put badge
	buf := &bytes.Buffer{}
	err = GenerateAndSaveBadge(buf, Config{Badge: Badge{CDN: cdn}}, coverage)
	require.NoError(t, err)
	assert.Contains(t, buf.String(), "Badge with updated coverage uploaded to CDN.")
	assertS3HasBadge(t, s3Client, cdn, coverage)

	// put badge again - no change
	buf = &bytes.Buffer{}
	err = GenerateAndSaveBadge(buf, Config{Badge: Badge{CDN: cdn}}, coverage)
	require.NoError(t, err)
	assert.Contains(t, buf.String(), "Badge with same coverage already uploaded to CDN.")
	assertS3HasBadge(t, s3Client, cdn, coverage)

	// put badge again - expect change
	buf = &bytes.Buffer{}
	err = GenerateAndSaveBadge(buf, Config{Badge: Badge{CDN: cdn}}, 10)
	require.NoError(t, err)
	assert.Contains(t, buf.String(), "Badge with updated coverage uploaded to CDN.")
	assertS3HasBadge(t, s3Client, cdn, 10)
}

func Test_GenerateAndSaveBadge_SaveToBranch_NoAction(t *testing.T) {
	t.Parallel()

	if testing.Short() {
		return
	}

	const coverage = 100

	err := GenerateAndSaveBadge(nil,
		Config{
			Badge: Badge{
				Git: Git{
					Token:      `🔑`,
					Owner:      "owner",
					Repository: "repo",
				},
			},
		}, coverage)
	assert.Error(t, err)
}

func assertS3HasBadge(t *testing.T, s3Client *s3.S3, cdn CDN, coverage int) {
	t.Helper()

	res, err := s3Client.GetObject(&s3.GetObjectInput{
		Bucket: &cdn.BucketName,
		Key:    &cdn.FileName,
	})
	require.NoError(t, err)

	resData, err := io.ReadAll(res.Body)
	assert.NoError(t, err)

	expectedData, err := badge.Generate(coverage)
	assert.NoError(t, err)
	assert.Equal(t, expectedData, resData)
}
