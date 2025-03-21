package coverage_test

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"golang.org/x/tools/cover"

	. "github.com/vladopajic/go-test-coverage/v2/pkg/testcoverage/coverage"
	"github.com/vladopajic/go-test-coverage/v2/pkg/testcoverage/path"
	"github.com/vladopajic/go-test-coverage/v2/pkg/testcoverage/testdata"
)

const (
	testdataDir             = "../testdata/"
	profileOK               = testdataDir + testdata.ProfileOK
	profileOKFull           = testdataDir + testdata.ProfileOKFull
	profileOKNoBadge        = testdataDir + testdata.ProfileOKNoBadge
	profileOKNoStatements   = testdataDir + testdata.ProfileOKNoStatements
	profileNOK              = testdataDir + testdata.ProfileNOK
	profileNOKInvalidLength = testdataDir + testdata.ProfileNOKInvalidLength
	profileNOKInvalidData   = testdataDir + testdata.ProfileNOKInvalidData

	prefix        = "github.com/vladopajic/go-test-coverage/v2"
	coverFilename = "pkg/testcoverage/coverage/cover.go"

	sourceDir = "../../../"
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
	stats, err = GenerateCoverageStats(Config{
		Profiles:  []string{profileNOK},
		SourceDir: sourceDir,
	})
	assert.Error(t, err)
	assert.Empty(t, stats)

	// should be okay to read valid profile
	stats1, err := GenerateCoverageStats(Config{
		Profiles:  []string{profileOK},
		SourceDir: sourceDir,
	})
	assert.NoError(t, err)
	assert.NotEmpty(t, stats1)

	// should be okay to read valid profile
	stats2, err := GenerateCoverageStats(Config{
		Profiles:     []string{profileOK},
		ExcludePaths: []string{`cover\.go$`},
		SourceDir:    sourceDir,
	})
	assert.NoError(t, err)
	assert.NotEmpty(t, stats2)
	// stats2 should have less total statements because cover.go should have been excluded
	assert.Greater(t, StatsCalcTotal(stats1).Total, StatsCalcTotal(stats2).Total)

	// should have total coverage because of second profile
	stats3, err := GenerateCoverageStats(Config{
		Profiles:  []string{profileOK, profileOKFull},
		SourceDir: sourceDir,
	})
	assert.NoError(t, err)
	assert.NotEmpty(t, stats3)
	assert.Equal(t, 100, StatsCalcTotal(stats3).CoveredPercentage())

	// should not have `badge/generate.go` in statistics because it has no statements
	stats4, err := GenerateCoverageStats(Config{
		Profiles:  []string{profileOKNoStatements},
		SourceDir: sourceDir,
	})
	assert.NoError(t, err)
	assert.Len(t, stats4, 1)
	assert.NotContains(t, `badge/generate.go`, stats4[0].Name)
}

func Test_findFile(t *testing.T) {
	t.Parallel()

	if testing.Short() {
		return
	}

	const filename = "pkg/testcoverage/coverage/cover.go"

	findFile := FindFileCreator("../../../")
	findFileFallbackToImport := FindFileCreator("")

	file, noPrefixName, found := findFile(prefix + "/" + filename)
	assert.True(t, found)
	assert.Equal(t, filename, noPrefixName)
	assert.True(t, strings.HasSuffix(file, path.NormalizeForOS(filename)))

	file, noPrefixName, found = findFileFallbackToImport(prefix + "/" + filename)
	assert.True(t, found)
	assert.Equal(t, filename, noPrefixName)
	assert.True(t, strings.HasSuffix(file, path.NormalizeForOS(filename)))

	_, _, found = findFile(prefix + "/main1.go")
	assert.False(t, found)

	_, _, found = findFile("")
	assert.False(t, found)

	_, _, found = findFile(prefix)
	assert.False(t, found)
}

func Test_findAnnotations(t *testing.T) {
	t.Parallel()

	_, err := FindAnnotations(nil)
	assert.Error(t, err)

	_, err = FindAnnotations([]byte{})
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

	comments, err := FindAnnotations([]byte(source))
	assert.NoError(t, err)
	assert.Equal(t, []int{3, 5}, pluckStartLine(comments))
}

func Test_findFuncs(t *testing.T) {
	t.Parallel()

	_, _, err := FindFuncsAndBlocks(nil)
	assert.Error(t, err)

	_, _, err = FindFuncsAndBlocks([]byte{})
	assert.Error(t, err)

	const source = `
	package foo
	func foo() int {
		return 1
	}
	func bar() int {
		a := 0
		for range 10 {
			a += 1
		}
		return a
	}
	func baraba() int {
		a := 0
		for i:=0;i<10; i++ {
			a += 1
		}
		return a
	}
	func zab(a int) int {
		if a == 0 {
			return a + 1
		} else if a == 1 {
			return a + 2
		}
		return a
	}
	`

	funcs, blocks, err := FindFuncsAndBlocks([]byte(source))
	assert.NoError(t, err)
	assert.Equal(t, []int{3, 6, 13, 20}, pluckStartLine(funcs))
	assert.Equal(t, []Extent{
		{8, 16, 10, 4},
		{15, 22, 17, 4},
		{21, 13, 23, 4},
		{23, 20, 25, 4},
	}, blocks)
}

func Test_sumCoverage(t *testing.T) {
	t.Parallel()

	funcs := []Extent{
		{StartLine: 1, EndLine: 10},
		{StartLine: 12, EndLine: 20},
	}
	profile := &cover.Profile{Blocks: []cover.ProfileBlock{
		{StartLine: 1, EndLine: 2, NumStmt: 1},
		{StartLine: 2, EndLine: 3, NumStmt: 1},
		{StartLine: 4, EndLine: 5, NumStmt: 1},
		{StartLine: 5, EndLine: 6, NumStmt: 1},
		{StartLine: 6, EndLine: 10, NumStmt: 1},
		{StartLine: 12, EndLine: 20, NumStmt: 5},
	}}

	s := SumCoverage(profile, funcs, nil, nil)
	expected := Stats{Total: 10, Covered: 0, UncoveredLines: []int{
		1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 12, 13, 14, 15, 16, 17, 18, 19, 20,
	}}
	assert.Equal(t, expected, s)

	// Coverage should be empty when every function is excluded
	s = SumCoverage(profile, funcs, nil, funcs)
	assert.Equal(t, Stats{Total: 0, Covered: 0}, s)

	// Case when annotations is set on block (it should ignore whole block)
	annotations := []Extent{{StartLine: 4, EndLine: 4}}
	blocks := []Extent{{StartLine: 4, EndLine: 10}}
	s = SumCoverage(profile, funcs, blocks, annotations)
	expected = Stats{Total: 7, Covered: 0, UncoveredLines: []int{
		1, 2, 3, 12, 13, 14, 15, 16, 17, 18, 19, 20,
	}}
	assert.Equal(t, expected, s)
}

func pluckStartLine(extents []Extent) []int {
	res := make([]int, len(extents))
	for i, e := range extents {
		res[i] = e.StartLine
	}

	return res
}
