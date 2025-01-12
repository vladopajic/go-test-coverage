package badgestorer_test

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"

	. "github.com/subhambhardwaj/go-test-coverage/v2/pkg/testcoverage/badgestorer"
)

func Test_File(t *testing.T) {
	t.Parallel()

	if testing.Short() {
		return
	}

	data := []byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 0}

	t.Run("invalid file", func(t *testing.T) {
		t.Parallel()

		s := NewFile(t.TempDir())
		updated, err := s.Store(data)
		assert.Error(t, err) // should not be able to write to directory
		assert.False(t, updated)
	})

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		testFile := t.TempDir() + "/badge.svg"

		s := NewFile(testFile)
		updated, err := s.Store(data)
		assert.NoError(t, err)
		assert.True(t, updated)

		contentBytes, err := os.ReadFile(testFile)
		assert.NoError(t, err)
		assert.Equal(t, data, contentBytes)
	})
}
