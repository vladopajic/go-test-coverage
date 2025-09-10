package coverage

import (
	"bytes"
	"errors"
	"fmt"
	"math"
	"regexp"
	"strconv"
	"strings"
)

type Extent struct {
	StartLine int
	StartCol  int
	EndLine   int
	EndCol    int
}

type Stats struct {
	Name                       string
	Total                      int64
	Covered                    int64
	Threshold                  int
	UncoveredLines             []int
	AnnotationsWithoutComments []Extent
}

func (s Stats) UncoveredLinesCount() int {
	return int(s.Total - s.Covered)
}

func (s Stats) CoveredPercentage() int {
	return CoveredPercentage(s.Total, s.Covered)
}

func (s Stats) CoveredPercentageF() float64 {
	return coveredPercentageF(s.Total, s.Covered, true)
}

func (s Stats) CoveredPercentageFNR() float64 {
	return coveredPercentageF(s.Total, s.Covered, false)
}

//nolint:mnd // relax
func (s Stats) Str() string {
	c := s.CoveredPercentage()

	if c == 100 { // precision not needed
		return fmt.Sprintf("%d%% (%d/%d)", c, s.Covered, s.Total)
	} else if c < 10 { // adds space for singe digit number
		return fmt.Sprintf(" %.1f%% (%d/%d)", s.CoveredPercentageF(), s.Covered, s.Total)
	}

	return fmt.Sprintf("%.1f%% (%d/%d)", s.CoveredPercentageF(), s.Covered, s.Total)
}

func StatsSearchMap(stats []Stats) map[string]Stats {
	m := make(map[string]Stats)
	for _, s := range stats {
		m[s.Name] = s
	}

	return m
}

func CoveredPercentage(total, covered int64) int {
	return int(coveredPercentageF(total, covered, true))
}

//nolint:mnd // relax
func coveredPercentageF(total, covered int64, round bool) float64 {
	if total == 0 {
		return 0
	}

	if covered == total {
		return 100
	}

	p := float64(covered*100) / float64(total)

	if !round {
		return p
	}

	// round to %.1f
	return float64(int(math.Round(p*10))) / 10
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

func StatsCalcTotal(stats []Stats) Stats {
	total := Stats{}

	for _, s := range stats {
		total.Total += s.Total
		total.Covered += s.Covered
	}

	return total
}

func StatsPluckName(stats []Stats) []string {
	result := make([]string, len(stats))

	for i, s := range stats {
		result[i] = s.Name
	}

	return result
}

func StatsFilterWithUncoveredLines(stats []Stats) []Stats {
	return filter(stats, func(s Stats) bool {
		return len(s.UncoveredLines) > 0
	})
}

func StatsFilterWithCoveredLines(stats []Stats) []Stats {
	return filter(stats, func(s Stats) bool {
		return len(s.UncoveredLines) == 0
	})
}

// StatsFilterWithMissingExplanations returns stats that have missing explanations
func StatsFilterWithMissingExplanations(stats []Stats) []Stats {
	return filter(stats, func(s Stats) bool {
		return len(s.AnnotationsWithoutComments) > 0
	})
}

func filter[T any](slice []T, predicate func(T) bool) []T {
	var result []T

	for _, value := range slice {
		if predicate(value) {
			result = append(result, value)
		}
	}

	return result
}

func StatsSerialize(stats []Stats) []byte {
	b := bytes.Buffer{}
	sep, nl := []byte(";"), []byte("\n")

	//nolint:errcheck // relax
	for _, s := range stats {
		b.WriteString(s.Name)
		b.Write(sep)
		b.WriteString(strconv.FormatInt(s.Total, 10))
		b.Write(sep)
		b.WriteString(strconv.FormatInt(s.Covered, 10))
		b.Write(nl)
	}

	return b.Bytes()
}

var ErrInvalidFormat = errors.New("invalid format")

func StatsDeserialize(b []byte) ([]Stats, error) {
	deserializeLine := func(bl []byte) (Stats, error) {
		fields := bytes.Split(bl, []byte(";"))
		if len(fields) != 3 { //nolint:mnd // relax
			return Stats{}, ErrInvalidFormat
		}

		t, err := strconv.ParseInt(strings.TrimSpace(string(fields[1])), 10, 64)
		if err != nil {
			return Stats{}, ErrInvalidFormat
		}

		c, err := strconv.ParseInt(strings.TrimSpace(string(fields[2])), 10, 64)
		if err != nil {
			return Stats{}, ErrInvalidFormat
		}

		return Stats{
			Name:    strings.TrimSpace(string(fields[0])),
			Total:   t,
			Covered: c,
		}, nil
	}

	lines := bytes.Split(b, []byte("\n"))
	result := make([]Stats, 0, len(lines))

	for _, l := range lines {
		if len(l) == 0 {
			continue
		}

		s, err := deserializeLine(l)
		if err != nil {
			return nil, err
		}

		result = append(result, s)
	}

	return result, nil
}
