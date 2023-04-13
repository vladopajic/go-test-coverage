package testcoverage

import (
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
)

//nolint:gochecknoglobals // relax
var separatorToReplace = regexp.QuoteMeta(string(filepath.Separator))

func normalizePathInRegex(path string) string {
	if runtime.GOOS != "windows" {
		return path
	}

	clean := regexp.MustCompile(`\\+/`).
		ReplaceAllStringFunc(path, func(s string) string {
			if strings.Count(s, "\\")%2 == 0 {
				return s
			}
			return s[1:]
		})

	return strings.ReplaceAll(clean, "/", separatorToReplace)
}

func matches(regexps []*regexp.Regexp, str string) bool {
	for _, r := range regexps {
		if r.MatchString(str) {
			return true
		}
	}

	return false
}
