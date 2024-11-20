package coverage

import (
	"bytes"
	"errors"
	"fmt"
	"regexp"
	"strconv"
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

func CalcTotalStats(stats []Stats) Stats {
	total := Stats{}

	for _, s := range stats {
		total.Total += s.Total
		total.Covered += s.Covered
	}

	return total
}

func SerializeStats(stats []Stats) []byte {
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

func DeserializeStats(b []byte) ([]Stats, error) {
	deserializeLine := func(bl []byte) (Stats, error) {
		fields := bytes.Split(bl, []byte(";"))
		if len(fields) != 3 { //nolint:mnd // relax
			return Stats{}, ErrInvalidFormat
		}

		t, err := strconv.ParseInt(string(fields[1]), 10, 64)
		if err != nil {
			return Stats{}, ErrInvalidFormat
		}

		c, err := strconv.ParseInt(string(fields[2]), 10, 64)
		if err != nil {
			return Stats{}, ErrInvalidFormat
		}

		return Stats{Name: string(fields[0]), Total: t, Covered: c}, nil
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
