package coverage

import (
	"fmt"
	"regexp"
	"strings"
)

type Stats struct {
	Name      string
	Total     int64
	Covered   int64
	Threshold int
}

func (s Stats) CoveredPercentage() int {
	return CoveredPercentage(s.Total, s.Covered)
}

//nolint:mnd // relax
func (s Stats) Str() string {
	c := s.CoveredPercentage()

	if c == 100 { // precision not needed
		return fmt.Sprintf("%d%% (%d/%d)", c, s.Covered, s.Total)
	} else if c < 10 { // adds space for singe digit number
		return fmt.Sprintf(" %.1f%% (%d/%d)", coveredPercentageF(s.Total, s.Covered), s.Covered, s.Total)
	}

	return fmt.Sprintf("%.1f%% (%d/%d)", coveredPercentageF(s.Total, s.Covered), s.Covered, s.Total)
}

func CoveredPercentage(total, covered int64) int {
	return int(coveredPercentageF(total, covered))
}

//nolint:mnd // relax
func coveredPercentageF(total, covered int64) float64 {
	if total == 0 {
		return 0
	}

	if covered == total {
		return 100
	}

	return float64(covered*100) / float64(total)
}

func stripPrefix(name, prefix string) string {
	if prefix == "" {
		return name
	}

	if string(prefix[len(prefix)-1]) != "/" {
		prefix += "/"
	}

	return strings.Replace(name, prefix, "", 1)
}

func matches(regexps []*regexp.Regexp, str string) bool {
	for _, r := range regexps {
		if r.MatchString(str) {
			return true
		}
	}

	return false
}

func compileExcludePathRules(excludePaths []string) []*regexp.Regexp {
	if len(excludePaths) == 0 {
		return nil
	}

	compiled := make([]*regexp.Regexp, len(excludePaths))

	for i, pattern := range excludePaths {
		compiled[i] = regexp.MustCompile(pattern)
	}

	return compiled
}

func CalcTotalStats(coverageStats []Stats) Stats {
	totalStats := Stats{}

	for _, stats := range coverageStats {
		totalStats.Total += stats.Total
		totalStats.Covered += stats.Covered
	}

	return totalStats
}
