package coverage

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCoverageForFile_InvalidSource(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	require.NoError(t, os.WriteFile(
		filepath.Join(dir, "go.mod"),
		[]byte("module example.com/test\n"),
		0o600,
	))
	require.NoError(t, os.WriteFile(
		filepath.Join(dir, "invalid.go"),
		[]byte("package"),
		0o600,
	))

	profile := filepath.Join(dir, "cover.profile")
	require.NoError(t, os.WriteFile(
		profile,
		[]byte(`mode: atomic
example.com/test/invalid.go:1.1,1.1 1 0
`),
		0o600))

	_, err := GenerateCoverageStats(Config{Profiles: []string{profile}, SourceDir: dir})
	assert.Error(t, err)
}

func TestGenerateCoverageStats_OnlyIgnoredStatements(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	require.NoError(t, os.WriteFile(
		filepath.Join(dir, "go.mod"),
		[]byte("module example.com/test\n"),
		0o600,
	))
	require.NoError(t, os.WriteFile(
		filepath.Join(dir, "ignored.go"),
		[]byte(`package test
func ignored() { // coverage-ignore
}
`),
		0o600))

	profile := filepath.Join(dir, "cover.profile")
	require.NoError(t, os.WriteFile(
		profile,
		[]byte(`mode: atomic
example.com/test/ignored.go:2.1,3.1 1 0
`),
		0o600))

	stats, err := GenerateCoverageStats(Config{Profiles: []string{profile}, SourceDir: dir})
	require.NoError(t, err)
	assert.Empty(t, stats)
}

func TestGenerateCoverageStats_TrivialErrorsIgnored(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	require.NoError(t, os.WriteFile(
		filepath.Join(dir, "go.mod"),
		[]byte("module example.com/test\n"),
		0o600,
	))
	require.NoError(t, os.WriteFile(
		filepath.Join(dir, "trivial.go"),
		[]byte(`package test
import "errors"
func ignored() (interface{}, error) {
	err := foo()
	if err != nil {
		return nil, err
	}
	if err := foo(); err != nil {
		return nil, err
	}
	val := bar()
	if val == 0 {
		return nil, errors.New("...")
	}
	return nil, nil
}
`),
		0o600))

	profile := filepath.Join(dir, "cover.profile")
	require.NoError(t, os.WriteFile(
		profile,
		[]byte(`mode: atomic
example.com/test/trivial.go:4.2,5.16 2 1
example.com/test/trivial.go:6.3,7.1 1 0
example.com/test/trivial.go:8.2,8.30 1 1
example.com/test/trivial.go:9.3,10.1 1 0
example.com/test/trivial.go:11.2,12.14 2 1
example.com/test/trivial.go:13.3,14.1 1 0
example.com/test/trivial.go:15.2,15.17 1 1
`),
		0o600))

	stats, err := GenerateCoverageStats(Config{
		Profiles:                  []string{profile},
		SourceDir:                 dir,
		ExcludeTrivialErrorChecks: true,
	})
	require.NoError(t, err)
	require.Len(t, stats, 1)

	expected := Stats{
		Name:                       "trivial.go",
		Total:                      7,
		Covered:                    6,
		Threshold:                  0,
		UncoveredLines:             []int{13, 14},
		AnnotationsWithoutComments: []int{},
	}
	assert.Equal(t, expected, stats[0])
}
