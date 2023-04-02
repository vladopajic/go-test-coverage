package testcoverage_test

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"

	. "github.com/vladopajic/go-test-coverage/pkg/testcoverage"
)

const (
	profileOK  = "testdata/ok.profile"
	profileNOK = "testdata/nok.profile"
)

func TestCheck(t *testing.T) {
	t.Parallel()

	if testing.Short() {
		return
	}

	t.Run("no profile", func(t *testing.T) {
		t.Parallel()

		buf := &bytes.Buffer{}
		result, err := Check(buf, Config{})
		assert.Error(t, err)
		assert.Empty(t, result)
	})

	t.Run("ok pass", func(t *testing.T) {
		t.Parallel()

		buf := &bytes.Buffer{}
		cfg := Config{Profile: profileOK, Threshold: Threshold{Total: 65}}
		result, err := Check(buf, cfg)
		assert.NoError(t, err)
		assert.True(t, result.Pass())
	})

	t.Run("ok fail", func(t *testing.T) {
		t.Parallel()

		buf := &bytes.Buffer{}
		cfg := Config{Profile: profileOK, Threshold: Threshold{Total: 100}}
		result, err := Check(buf, cfg)
		assert.NoError(t, err)
		assert.False(t, result.Pass())
	})

	t.Run("nok", func(t *testing.T) {
		t.Parallel()

		buf := &bytes.Buffer{}
		cfg := Config{Profile: profileNOK, Threshold: Threshold{Total: 65}}
		result, err := Check(buf, cfg)
		assert.Error(t, err)
		assert.False(t, result.Pass())
	})
}
