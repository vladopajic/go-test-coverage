package path

import (
	"path/filepath"
	"runtime"
	"strings"
)

const separatorToReplace = string(filepath.Separator)

func NormalizeForOS(path string) string {
	if runtime.GOOS != "windows" {
		return path
	}

	return strings.ReplaceAll(path, "/", separatorToReplace)
}

func NormalizeForTool(path string) string {
	if runtime.GOOS != "windows" {
		return path
	}

	return strings.ReplaceAll(path, separatorToReplace, "/")
}
