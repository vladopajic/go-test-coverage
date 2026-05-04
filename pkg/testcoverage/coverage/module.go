package coverage

import (
	"bufio"
	"os"
	"path/filepath"
	"strings"

	"github.com/vladopajic/go-test-coverage/v2/pkg/testcoverage/logger"
)

//nolint:nonamedreturns // relax
func findModuleDirective(rootDir string) (module string, dir string) {
	logger.L.Debug().Str("root dir", rootDir).Msg("searching for go.mod")

	goModFile := findGoModFile(rootDir)
	if goModFile == "" {
		logger.L.Warn().Str("dir", rootDir).
			Msg("go.mod file not found in root directory (consider setting up source dir)")

		return "", rootDir
	}

	logger.L.Debug().Str("file", goModFile).Msg("go.mod file found")

	module = readModuleDirective(goModFile)
	if module == "" { // coverage-ignore
		logger.L.Warn().Msg("`module` directive not found")
	}

	dir = filepath.Dir(goModFile)

	logger.L.Debug().Str("module", module).Msg("using module directive")
	logger.L.Debug().Str("rootdir", dir).Msg("root dir")

	return module, dir
}

func findGoModFile(rootDir string) string {
	goModFile := findGoModFromRoot(rootDir)
	if goModFile != "" {
		return goModFile
	}

	// fallback: search the entire directory tree for go.mod
	return findGoModWithWalk(rootDir)
}

func findGoModWithWalk(rootDir string) string { // coverage-ignore
	var goModFiles []string

	err := filepath.Walk(rootDir, func(file string, info os.FileInfo, err error) error {
		if err != nil { // coverage-ignore
			return err
		}

		if info.Name() == "go.mod" {
			goModFiles = append(goModFiles, file)
		}

		return nil
	})
	if err != nil {
		logger.L.Error().Err(err).Msg("listing files (go.mod search)")
	}

	if len(goModFiles) == 0 {
		logger.L.Warn().Msg("go.mod file not found via walk method")
		return ""
	}

	if len(goModFiles) > 1 {
		logger.L.Warn().Msg("found multiple go.mod files via walk method")
		return ""
	}

	return goModFiles[0]
}

func findGoModFromRoot(rootDir string) string {
	files, err := os.ReadDir(rootDir)
	if err != nil { // coverage-ignore
		logger.L.Error().Err(err).Msg("reading directory")
		return ""
	}

	for _, info := range files {
		if info.Name() == "go.mod" {
			return filepath.Join(rootDir, info.Name())
		}
	}

	return ""
}

func readModuleDirective(filename string) string {
	file, err := os.Open(filename)
	if err != nil { // coverage-ignore
		return ""
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if strings.HasPrefix(line, "module ") {
			return strings.TrimSpace(strings.TrimPrefix(line, "module "))
		}
	}

	return "" // coverage-ignore
}
