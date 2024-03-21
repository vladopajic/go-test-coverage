package coverage_test

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"golang.org/x/tools/cover"

	. "github.com/vladopajic/go-test-coverage/v2/pkg/testcoverage/coverage"
	"github.com/vladopajic/go-test-coverage/v2/pkg/testcoverage/testdata"
)

const (
	testdataDir             = "../testdata/"
	profileOK               = testdataDir + testdata.ProfileOK
	profileOKFull           = testdataDir + testdata.ProfileOKFull
	profileOKNoPath         = testdataDir + testdata.ProfileOKNoPath
	profileNOK              = testdataDir + testdata.ProfileNOK
	profileNOKInvalidLength = testdataDir + testdata.ProfileNOKInvalidLength
	profileNOKInvalidData   = testdataDir + testdata.ProfileNOKInvalidData

	prefix        = "github.com/vladopajic/go-test-coverage/v2"
	coverFilename = "pkg/testcoverage/coverage/cover.go"
)

func Test_GenerateCoverageStats(t *testing.T) {
	t.Parallel()

	if testing.Short() {
		return
	}

	// should not be able to read directory
	stats, err := GenerateCoverageStats(Config{Profiles: []string{t.TempDir()}})
	assert.Error(t, err)
	assert.Empty(t, stats)

	// should get error parsing invalid profile file
	stats, err = GenerateCoverageStats(Config{Profiles: []string{profileNOK}})
	assert.Error(t, err)
	assert.Empty(t, stats)

	// should be okay to read valid profile
	stats1, err := GenerateCoverageStats(Config{Profiles: []string{profileOK}})
	assert.NoError(t, err)
	assert.NotEmpty(t, stats1)

	// should be okay to read valid profile
	stats2, err := GenerateCoverageStats(Config{
		Profiles:     []string{profileOK},
		ExcludePaths: []string{`cover\.go$`},
	})
	assert.NoError(t, err)
	assert.NotEmpty(t, stats2)
	// stats2 should have less total statements because cover.go should have been excluded
	assert.Greater(t, CalcTotalStats(stats1).Total, CalcTotalStats(stats2).Total)

	// should remove prefix from stats
	stats3, err := GenerateCoverageStats(Config{
		Profiles:    []string{profileOK},
		LocalPrefix: prefix,
	})
	assert.NoError(t, err)
	assert.NotEmpty(t, stats3)
	assert.Equal(t, CalcTotalStats(stats1), CalcTotalStats(stats3))
	assert.NotContains(t, stats3[0].Name, prefix)
	assert.NotEqual(t, 100, CalcTotalStats(stats3).CoveredPercentage())

	// should have total coverage because of second profle
	stats4, err := GenerateCoverageStats(Config{
		Profiles: []string{profileOK, profileOKFull},
	})
	assert.NoError(t, err)
	assert.NotEmpty(t, stats4)
	assert.Equal(t, 100, CalcTotalStats(stats4).CoveredPercentage())
}

func Test_findFile(t *testing.T) {
	t.Parallel()

	if testing.Short() {
		return
	}

	const filename = "pkg/testcoverage/coverage/cover.go"

	file, noPrefixName, err := FindFile(prefix+"/"+filename, "")
	assert.NoError(t, err)
	assert.Equal(t, filename, noPrefixName)
	assert.True(t, strings.HasSuffix(file, filename))

	file, noPrefixName, err = FindFile(prefix+"/"+filename, prefix)
	assert.NoError(t, err)
	assert.Equal(t, filename, noPrefixName)
	assert.True(t, strings.HasSuffix(file, filename))

	_, _, err = FindFile(prefix+"/main1.go", "")
	assert.Error(t, err)

	_, _, err = FindFile("", "")
	assert.Error(t, err)

	_, _, err = FindFile(prefix, "")
	assert.Error(t, err)
}

func Test_findComments(t *testing.T) {
	t.Parallel()

	_, err := FindComments(nil)
	assert.Error(t, err)

	_, err = FindComments([]byte{})
	assert.Error(t, err)

	const source = `
		package foo

		func foo() int { // coverage-ignore
			a := 0
			for i := range 10 { // coverage-ignore
				a += i
			}
			return a
		}
	`
	comments, err := FindComments([]byte(source))
	assert.NoError(t, err)
	assert.Equal(t, []int{4, 6}, pluckStartLine(comments))
}

func Test_findFuncs(t *testing.T) {
	t.Parallel()

	_, err := FindFuncs(nil)
	assert.Error(t, err)

	_, err = FindFuncs([]byte{})
	assert.Error(t, err)

	const source = `
	package foo

	func foo() int {
		a := 0
		return a
	}

	func bar() int {
		return 1
	}
	`
	funcs, err := FindFuncs([]byte(source))
	assert.NoError(t, err)
	assert.Equal(t, []int{4, 9}, pluckStartLine(funcs))
}

func Test_coverageForFile(t *testing.T) {
	t.Parallel()

	extent := []Extent{
		{StartLine: 1, EndLine: 10},
		{StartLine: 12, EndLine: 20},
	}
	profile := &cover.Profile{Blocks: []cover.ProfileBlock{
		{StartLine: 1, EndLine: 10, NumStmt: 5},
		{StartLine: 12, EndLine: 20, NumStmt: 5},
	}}

	s := CoverageForFile(profile, extent, nil)
	assert.Equal(t, Stats{Total: 10, Covered: 0}, s)

	// Coverage should be empty when there every function is excluded
	s = CoverageForFile(profile, extent, extent)
	assert.Empty(t, s)
}

func pluckStartLine(extents []Extent) []int {
	res := make([]int, len(extents))
	for i, e := range extents {
		res[i] = e.StartLine
	}

	return res
}
