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

type regRule struct {
	reg       *regexp.Regexp
	threshold int
}

func matches(regexps []regRule, str string) (int, bool) {
	for _, r := range regexps {
		if r.reg.MatchString(str) {
			return r.threshold, true
		}
	}

	return 0, false
}

func compileExcludePathRules(cfg Config) []regRule {
	if len(cfg.Exclude.Paths) == 0 {
		return nil
	}

	compiled := make([]regRule, 0, len(cfg.Exclude.Paths))

	for _, pattern := range cfg.Exclude.Paths {
		pattern = normalizePathInRegex(pattern)
		compiled = append(compiled, regRule{
			reg: regexp.MustCompile(pattern),
		})
	}

	return compiled
}

func compileOverridePathRules(cfg Config) []regRule {
	if len(cfg.Override) == 0 {
		return nil
	}

	compiled := make([]regRule, 0, len(cfg.Override))

	for _, o := range cfg.Override {
		pattern := normalizePathInRegex(o.Path)
		compiled = append(compiled, regRule{
			reg:       regexp.MustCompile(pattern),
			threshold: o.Threshold,
		})
	}

	return compiled
}
