package coverage_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	. "github.com/vladopajic/go-test-coverage/v2/pkg/testcoverage/coverage"
	"github.com/vladopajic/go-test-coverage/v2/pkg/testcoverage/path"
)

func Test_FindGoModFile(t *testing.T) {
	t.Parallel()

	assert.Empty(t, FindGoModFile(""))
	assert.Equal(t, path.NormalizeForTool("../../../go.mod"), FindGoModFile("../../../"))
}
