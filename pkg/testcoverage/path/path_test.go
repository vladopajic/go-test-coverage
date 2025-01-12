package path_test

import (
	"runtime"
	"testing"

	. "github.com/subhambhardwaj/go-test-coverage/v2/pkg/testcoverage/path"

	"github.com/stretchr/testify/assert"
)

func Test_NormalizeForOS(t *testing.T) {
	t.Parallel()

	if runtime.GOOS == "windows" {
		assert.Equal(t, "foo\\bar", NormalizeForOS("foo/bar"))
	} else {
		assert.Equal(t, "foo/bar", NormalizeForOS("foo/bar"))
	}
}

func Test_NormalizeForTool(t *testing.T) {
	t.Parallel()

	if runtime.GOOS == "windows" {
		assert.Equal(t, "foo/bar", NormalizeForTool("foo\\bar"))
	} else {
		assert.Equal(t, "foo/bar", NormalizeForTool("foo/bar"))
	}
}
