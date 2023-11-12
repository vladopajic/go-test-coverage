package badgestorer_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	. "github.com/vladopajic/go-test-coverage/v2/pkg/testcoverage/badgestorer"
)

func Test_Github_Error(t *testing.T) {
	t.Parallel()

	if testing.Short() {
		return
	}

	data := []byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 0}
	cfg := Git{
		Token:      `🔑`,
		Owner:      "owner",
		Repository: "repo",
	}
	s := NewGithub(cfg)

	updated, err := s.Store(data)
	assert.Error(t, err)
	assert.False(t, updated)
}
