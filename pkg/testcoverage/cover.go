package testcoverage

import (
	"fmt"
	"go/ast"
	"go/build"
	"go/parser"
	"go/token"
	"math"
	"path/filepath"
	"strings"

	"golang.org/x/tools/cover"
)

//nolint:wrapcheck // relax
func GenerateCoverageStats(profileFileName string) ([]CoverageStats, error) {
	profiles, err := cover.ParseProfiles(profileFileName)
	if err != nil {
		return nil, err
	}

	fileStats := make([]CoverageStats, 0, len(profiles))

	for _, profile := range profiles {
		file, err := findFile(profile.FileName)
		if err != nil {
			return nil, err
		}

		funcs, err := findFuncs(file)
		if err != nil {
			return nil, err
		}

		s := CoverageStats{
			name: profile.FileName,
		}

		for _, f := range funcs {
			c, t := f.coverage(profile)
			s.total += t
			s.covered += c
		}

		fileStats = append(fileStats, s)
	}

	return fileStats, nil
}

type CoverageStats struct {
	name    string
	total   int64
	covered int64
}

func (s *CoverageStats) CoveredPercentage() int {
	if s.total == 0 {
		return 0
	}

	//nolint:gomnd // relax
	return int(math.Round((float64(s.covered*100) / float64(s.total))))
}

// findFuncs parses the file and returns a slice of FuncExtent descriptors.
func findFuncs(name string) ([]*FuncExtent, error) {
	fset := token.NewFileSet()

	parsedFile, err := parser.ParseFile(fset, name, nil, 0)
	if err != nil {
		return nil, err //nolint:wrapcheck // relax
	}

	visitor := &FuncVisitor{
		fset: fset,
		name: name,
	}
	ast.Walk(visitor, parsedFile)

	return visitor.funcs, nil
}

// FuncExtent describes a function's extent in the source by file and position.
type FuncExtent struct {
	name      string
	startLine int
	startCol  int
	endLine   int
	endCol    int
}

// FuncVisitor implements the visitor that builds the function position list for a file.
type FuncVisitor struct {
	fset  *token.FileSet
	name  string
	funcs []*FuncExtent
}

// Visit implements the ast.Visitor interface.
func (v *FuncVisitor) Visit(node ast.Node) ast.Visitor {
	if n, ok := node.(*ast.FuncDecl); ok {
		start := v.fset.Position(n.Pos())
		end := v.fset.Position(n.End())
		fe := &FuncExtent{
			name:      n.Name.Name,
			startLine: start.Line,
			startCol:  start.Column,
			endLine:   end.Line,
			endCol:    end.Column,
		}
		v.funcs = append(v.funcs, fe)
	}

	return v
}

// coverage returns the fraction of the statements in the
// function that were covered, as a numerator and denominator.
//
// We could avoid making this n^2 overall by doing
// a single scan and annotating the functions, but the sizes of the data
// structures is never very large and the scan is almost instantaneous.
func (f *FuncExtent) coverage(profile *cover.Profile) (int64, int64) {
	var covered, total int64

	// The blocks are sorted, so we can stop counting as soon as
	// we reach the end of the relevant block.
	for _, b := range profile.Blocks {
		if b.StartLine > f.endLine || (b.StartLine == f.endLine && b.StartCol >= f.endCol) {
			// Past the end of the function.
			break
		}

		if b.EndLine < f.startLine || (b.EndLine == f.startLine && b.EndCol <= f.startCol) {
			// Before the beginning of the function
			continue
		}

		total += int64(b.NumStmt)

		if b.Count > 0 {
			covered += int64(b.NumStmt)
		}
	}

	return covered, total
}

// findFile finds the location of the named file in GOROOT, GOPATH etc.
func findFile(file string) (string, error) {
	dir, file := filepath.Split(file)

	pkg, err := build.Import(dir, ".", build.FindOnly)
	if err != nil {
		return "", fmt.Errorf("can't find %q: %w", file, err)
	}

	return filepath.Join(pkg.Dir, file), nil
}

func checkCoverageStatsBelowThreshold(
	coverageStats []CoverageStats,
	threshold int,
) []CoverageStats {
	belowThreshold := make([]CoverageStats, 0)

	for _, stats := range coverageStats {
		if stats.CoveredPercentage() < threshold {
			belowThreshold = append(belowThreshold, stats)
		}
	}

	return belowThreshold
}

func calcTotalStats(coverageStats []CoverageStats) CoverageStats {
	totalStats := CoverageStats{}

	for _, stats := range coverageStats {
		totalStats.total += stats.total
		totalStats.covered += stats.covered
	}

	return totalStats
}

func makePackageStats(coverageStats []CoverageStats) []CoverageStats {
	packageStats := make(map[string]CoverageStats)

	for _, stats := range coverageStats {
		pkg := packageForFile(stats.name)

		var pkgStats CoverageStats
		if s, ok := packageStats[pkg]; ok {
			pkgStats = s
		} else {
			pkgStats = CoverageStats{name: pkg}
		}

		pkgStats.total += stats.total
		pkgStats.covered += stats.covered
		packageStats[pkg] = pkgStats
	}

	packageStatsSlice := make([]CoverageStats, 0, len(packageStats))
	for _, stats := range packageStats {
		packageStatsSlice = append(packageStatsSlice, stats)
	}

	return packageStatsSlice
}

func packageForFile(filename string) string {
	i := strings.LastIndex(filename, "/")
	if i == -1 {
		return filename
	}

	return filename[:i]
}
