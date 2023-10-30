package testcoverage_test

import (
	"os"
	"testing"

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

	err := GenerateAndSaveBadge(Config{
		Badge: Badge{
			CDN: CDN{Secret: "your-secrets-are-safu"},
		},
	}, 100)
	assert.Error(t, err)

	err = GenerateAndSaveBadge(Config{
		Badge: Badge{
			CDN: CDN{
				Key:        "key",
				Secret:     "pssst!",
				Region:     "eu-central-1",
				FileName:   "coverage.svg",
				BucketName: "badges",
			},
		},
	}, 100)
	assert.Error(t, err)
}
