package testcoverage_test

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"

	. "github.com/vladopajic/go-test-coverage/v2/pkg/testcoverage"
)

func Test_GenerateAndSaveBadge_SaveToFile(t *testing.T) {
	t.Parallel()

	if testing.Short() {
		return
	}

	testFile := t.TempDir() + "/badge.svg"

	{
		err := GenerateAndSaveBadge(Config{
			Badge: Badge{},
		}, 100)
		assert.NoError(t, err)

		contentBytes, err := os.ReadFile(testFile)
		assert.Error(t, err)
		assert.Empty(t, contentBytes)
	}

	{
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
}
