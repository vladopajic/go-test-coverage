package coverage

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

//nolint:forbidigo // relax
func findModuleDirective(rootDir string) string {
	goModFile := findGoModFile(rootDir)
	if goModFile == "" {
		fmt.Printf("could not find go.mod file in root dir: %s\n", rootDir)
		return ""
	}

	module := readModuleDirective(goModFile)
	if module == "" { // coverage-ignore
		fmt.Println("`module` directive not found")
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
