package coverage_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	. "github.com/vladopajic/go-test-coverage/v2/pkg/testcoverage/coverage"
)

func Test_FindGoModFile(t *testing.T) {
	t.Parallel()

	assert.Empty(t, FindGoModFile(""))
	assert.Equal(t, "../../../go.mod", FindGoModFile("../../../"))
}
