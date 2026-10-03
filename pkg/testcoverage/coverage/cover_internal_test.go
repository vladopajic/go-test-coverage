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
