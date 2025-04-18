package coverage

import (
	"bufio"
	"os"
	"path/filepath"
	"strings"

	"github.com/vladopajic/go-test-coverage/v2/pkg/testcoverage/logger"
)

func findModuleDirective(rootDir string) string {
	goModFile := findGoModFile(rootDir)
	if goModFile == "" {
		logger.L.Warn().Str("dir", rootDir).Msg("could not find go.mod file in root dir")
		return ""
	}

	module := readModuleDirective(goModFile)
	if module == "" { // coverage-ignore
		logger.L.Warn().Msg("`module` directive not found")
	}

	return module
}

func findGoModFile(rootDir string) string {
	var goModFile string

	//nolint:errcheck // error ignored because there is fallback mechanism for finding files
	filepath.Walk(rootDir, func(file string, info os.FileInfo, err error) error {
		if err != nil { // coverage-ignore
			return err
		}

		if info.Name() == "go.mod" {
			goModFile = file
			return filepath.SkipAll
		}

		return nil
	})

	return goModFile
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
