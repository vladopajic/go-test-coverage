package testcoverage

import (
	"regexp"
)

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

func compileOverridePathRules(cfg Config) []regRule {
	if len(cfg.Override) == 0 {
		return nil
	}

	compiled := make([]regRule, len(cfg.Override))

	for i, o := range cfg.Override {
		pattern := o.Path
		compiled[i] = regRule{
			reg:       regexp.MustCompile(pattern),
			threshold: o.Threshold,
		}
	}

	return compiled
}
