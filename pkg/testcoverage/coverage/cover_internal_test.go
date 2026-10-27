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

func TestGenerateCoverageStats_SingleReturnErrorsIgnored(t *testing.T) {
	if testing.Short() {
		return
	}

	t.Parallel()

	dir := t.TempDir()
	require.NoError(t, os.WriteFile(
		filepath.Join(dir, "go.mod"),
		[]byte("module example.com/test\n"),
		0o600,
	))
	require.NoError(t, os.WriteFile(
		filepath.Join(dir, "singleerror.go"),
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
func foo() error {
	return nil
}
func bar() int {
	return 0
}
`),
		0o600))

	profile := filepath.Join(dir, "cover.profile")
	require.NoError(t, os.WriteFile(
		profile,
		[]byte(`mode: atomic
example.com/test/singleerror.go:4.2,5.16 2 1
example.com/test/singleerror.go:6.3,7.1 1 0
example.com/test/singleerror.go:8.2,8.30 1 1
example.com/test/singleerror.go:9.3,10.1 1 0
example.com/test/singleerror.go:11.2,12.14 2 1
example.com/test/singleerror.go:13.3,14.1 1 0
example.com/test/singleerror.go:15.2,15.17 1 1
`),
		0o600))

	stats, err := GenerateCoverageStats(Config{
		Profiles:                       []string{profile},
		SourceDir:                      dir,
		ExcludeSingleReturnErrorChecks: true,
	})
	require.NoError(t, err)
	require.Len(t, stats, 1)

	expected := Stats{
		Name:                       "singleerror.go",
		Total:                      7,
		Covered:                    6,
		Threshold:                  0,
		UncoveredLines:             []int{13, 14},
		AnnotationsWithoutComments: []int{},
	}
	assert.Equal(t, expected, stats[0])
}

//nolint:paralleltest // This test loads packages from a temporary module.
func TestGenerateCoverageStats_SingleReturnErrorsLoadFailure(t *testing.T) {
	if testing.Short() {
		return
	}

	dir := t.TempDir()
	require.NoError(t, os.WriteFile(
		filepath.Join(dir, "go.mod"),
		[]byte("module example.com/test\n"),
		0o600,
	))
	require.NoError(t, os.WriteFile(
		filepath.Join(dir, "typeerror.go"),
		[]byte(`package test
var x int = "not an int"
`),
		0o600))

	profile := filepath.Join(dir, "cover.profile")
	require.NoError(t, os.WriteFile(
		profile,
		[]byte(`mode: atomic
example.com/test/typeerror.go:2.1,2.25 1 0
`),
		0o600))

	_, err := GenerateCoverageStats(Config{
		Profiles:                       []string{profile},
		SourceDir:                      dir,
		ExcludeSingleReturnErrorChecks: true,
	})
	require.ErrorContains(t, err, "loading package")
}

func TestGenerateCoverageStats_IndependentPackageLoads(t *testing.T) {
	t.Parallel()

	if testing.Short() {
		t.Skip("loads Go packages")
	}

	makeProject := func() (string, string) {
		dir := t.TempDir()
		require.NoError(t, os.WriteFile(
			filepath.Join(dir, "go.mod"),
			[]byte("module example.com/test\n\ngo 1.27\n"),
			0o600,
		))
		require.NoError(t, os.WriteFile(
			filepath.Join(dir, "a.go"),
			[]byte("package test\nfunc f(err error) error { if err != nil { return err }; return nil }\n"),
			0o600,
		))
		profile := filepath.Join(dir, "cover.out")
		require.NoError(t, os.WriteFile(
			profile,
			[]byte(`mode: set
example.com/test/a.go:2.25,2.41 1 1
example.com/test/a.go:2.41,2.55 1 0
example.com/test/a.go:2.57,2.67 1 1
`),
			0o600,
		))

		return dir, profile
	}
	dir1, profile1 := makeProject()
	dir2, profile2 := makeProject()
	broken := filepath.Join(dir1, "unrelated")
	require.NoError(t, os.Mkdir(broken, 0o700))
	require.NoError(t, os.WriteFile(
		filepath.Join(broken, "bad.go"),
		[]byte("package bad\nvar x int = \"bad\"\n"),
		0o600,
	))

	for _, project := range []struct{ dir, profile string }{{dir1, profile1}, {dir2, profile2}} {
		stats, err := GenerateCoverageStats(Config{
			Profiles:                       []string{project.profile},
			SourceDir:                      project.dir,
			ExcludeSingleReturnErrorChecks: true,
		})
		require.NoError(t, err)
		require.Len(t, stats, 1)
		assert.EqualValues(t, 2, stats[0].Total)
	}
}

func TestGenerateCoverageStats_RetryAfterLoadFailure(t *testing.T) {
	t.Parallel()

	if testing.Short() {
		t.Skip("loads Go packages")
	}

	dir := t.TempDir()
	require.NoError(t, os.WriteFile(
		filepath.Join(dir, "go.mod"),
		[]byte("module example.com/test\n\ngo 1.27\n"),
		0o600,
	))
	source := filepath.Join(dir, "a.go")
	require.NoError(t, os.WriteFile(
		source,
		[]byte("package test\nvar x int = \"invalid\"\nfunc f() {}\n"),
		0o600,
	))

	profile := filepath.Join(dir, "cover.out")
	require.NoError(t, os.WriteFile(
		profile,
		[]byte("mode: set\nexample.com/test/a.go:3.10,3.12 1 1\n"),
		0o600,
	))
	cfg := Config{Profiles: []string{profile}, SourceDir: dir, ExcludeSingleReturnErrorChecks: true}
	_, err := GenerateCoverageStats(cfg)
	require.Error(t, err)
	require.NoError(t, os.WriteFile(
		source,
		[]byte("package test\nvar x int = 1\nfunc f() {}\n"),
		0o600,
	))

	_, err = GenerateCoverageStats(cfg)
	require.NoError(t, err)
}

func TestGenerateCoverageStats_ExternalModule(t *testing.T) {
	t.Parallel()

	if testing.Short() {
		t.Skip("loads Go packages")
	}

	root := t.TempDir()
	app := filepath.Join(root, "app")
	dep := filepath.Join(root, "dep")

	require.NoError(t, os.Mkdir(app, 0o700))
	require.NoError(t, os.Mkdir(dep, 0o700))
	require.NoError(t, os.WriteFile(
		filepath.Join(app, "go.mod"),
		[]byte(`module example.com/app

go 1.27

require example.com/dep v0.0.0
replace example.com/dep => ../dep
`),
		0o600,
	))
	require.NoError(t, os.WriteFile(
		filepath.Join(dep, "go.mod"),
		[]byte("module example.com/dep\n\ngo 1.27\n"),
		0o600,
	))
	require.NoError(t, os.WriteFile(
		filepath.Join(dep, "a.go"),
		[]byte("package dep\nfunc F(err error) error { if err != nil { return err }; return nil }\n"),
		0o600,
	))
	profile := filepath.Join(app, "cover.out")
	require.NoError(t, os.WriteFile(
		profile,
		[]byte(`mode: set
example.com/dep/a.go:2.25,2.41 1 1
example.com/dep/a.go:2.41,2.55 1 0
example.com/dep/a.go:2.57,2.67 1 1
`),
		0o600,
	))
	stats, err := GenerateCoverageStats(Config{
		Profiles:                       []string{profile},
		SourceDir:                      app,
		ExcludeSingleReturnErrorChecks: true,
	})
	require.NoError(t, err)
	require.Len(t, stats, 1)
	assert.EqualValues(t, 2, stats[0].Total)
}

func TestTypedLoader_MissingPackageDirectory(t *testing.T) {
	t.Parallel()

	if testing.Short() {
		t.Skip("loads Go packages")
	}

	missingDir := filepath.Join(t.TempDir(), "missing")
	_, _, found := findDependencyFile("example.com/dep/file.go", missingDir)
	assert.False(t, found)

	_, err := findSingleReturnErrorChecks(filepath.Join(missingDir, "file.go"))
	require.ErrorContains(t, err, "loading package")
}

func TestTypedLoader_FileNotInPackage(t *testing.T) {
	t.Parallel()

	if testing.Short() {
		t.Skip("loads Go packages")
	}

	// Load this package successfully, but ask for a file it does not contain.
	dir, err := os.Getwd()
	require.NoError(t, err)

	_, err = findSingleReturnErrorChecks(filepath.Join(dir, "missing.go"))
	require.ErrorContains(t, err, "file not found in package")
}
