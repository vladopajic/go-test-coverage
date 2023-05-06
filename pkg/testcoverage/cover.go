package testcoverage

import (
	"fmt"
	"go/ast"
	"go/build"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"

	"golang.org/x/tools/cover"
)

func GenerateCoverageStats(cfg Config) ([]CoverageStats, error) {
	profiles, err := cover.ParseProfiles(cfg.Profile)
	if err != nil {
		return nil, fmt.Errorf("parsing profile file: %w", err)
	}

	fileStats := make([]CoverageStats, 0, len(profiles))
	excludeRules := compileExcludePathRules(cfg)

	for _, profile := range profiles {
		file, noPrefixName, err := findFile(profile.FileName, cfg.LocalPrefix)
		if err != nil {
			return nil, fmt.Errorf("could not find file [%s]: %w", profile.FileName, err)
		}

		if _, ok := matches(excludeRules, noPrefixName); ok {
			continue // this file is excluded
		}

		funcs, err := findFuncs(file)
		if err != nil {
			return nil, fmt.Errorf("failed parsing funcs from file [%s]: %w", profile.FileName, err)
		}

		s := CoverageStats{
			Name: noPrefixName,
		}

		for _, f := range funcs {
			c, t := f.coverage(profile)
			s.Total += t
			s.Covered += c
		}

		fileStats = append(fileStats, s)
	}

	return fileStats, nil
}

// findFile finds the location of the named file in GOROOT, GOPATH etc.
func findFile(file, prefix string) (string, string, error) {
	noPrefixName := stripPrefix(file, prefix)
	if _, err := os.Stat(noPrefixName); err == nil {
		return noPrefixName, noPrefixName, nil
	}

	dir, file := filepath.Split(file)

	pkg, err := build.Import(dir, ".", build.FindOnly)
	if err != nil {
		return "", "", fmt.Errorf("can't find %q: %w", file, err)
	}

	file = filepath.Join(pkg.Dir, file)
	noPrefixName = stripPrefix(file, pkg.Root)

	return file, noPrefixName, nil
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
