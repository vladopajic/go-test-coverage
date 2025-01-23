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

func compileOverridePathRules(cfg Config) ([]regRule, bool) {
	if len(cfg.Override) == 0 {
		return nil, false
	}

	compiled := make([]regRule, len(cfg.Override))

	for i, o := range cfg.Override {
		compiled[i] = regRule{
			reg:       regexp.MustCompile(o.Path),
			threshold: o.Threshold,
		}
	}

	return compiled, true
}
