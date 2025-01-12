package coverage_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	. "github.com/subhambhardwaj/go-test-coverage/v2/pkg/testcoverage/coverage"
)

func Test_parseProfiles(t *testing.T) {
	t.Parallel()

	if testing.Short() {
		return
	}

	_, err := ParseProfiles([]string{""})
	assert.Error(t, err)

	_, err = ParseProfiles([]string{profileOK, profileNOKInvalidLength})
	assert.Error(t, err)

	_, err = ParseProfiles([]string{profileOK, profileNOKInvalidData})
	assert.Error(t, err)

	p1, err := ParseProfiles([]string{profileOK, profileOKFull})
	assert.NoError(t, err)
	assert.NotEmpty(t, p1)

	p2, err := ParseProfiles([]string{profileOKFull})
	assert.NoError(t, err)
	assert.Equal(t, p1, p2)

	p3, err := ParseProfiles([]string{profileOK})
	assert.NoError(t, err)
	assert.NotEmpty(t, p3)

	p4, err := ParseProfiles([]string{profileOKNoBadge, profileOK})
	assert.NoError(t, err)
	assert.Equal(t, p3, p4)

	p5, err := ParseProfiles([]string{profileOK, profileOKNoBadge})
	assert.NoError(t, err)
	assert.Equal(t, p4, p5)
}
