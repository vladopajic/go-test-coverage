package coverage

import (
	"fmt"
	"go/ast"
	"go/build"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"

	"golang.org/x/tools/cover"
)

const IgnoreText = "coverage-ignore"

type Config struct {
	Profile      string
	LocalPrefix  string
	ExcludePaths []string
}

func GenerateCoverageStats(cfg Config) ([]Stats, error) {
	profiles, err := cover.ParseProfiles(cfg.Profile)
	if err != nil {
		return nil, fmt.Errorf("parsing profile file: %w", err)
	}

	fileStats := make([]Stats, 0, len(profiles))
	excludeRules := compileExcludePathRules(cfg.ExcludePaths)

	for _, profile := range profiles {
		file, noPrefixName, err := findFile(profile.FileName, cfg.LocalPrefix)
		if err != nil {
			return nil, fmt.Errorf("could not find file [%s]: %w", profile.FileName, err)
		}

		if ok := matches(excludeRules, noPrefixName); ok {
			continue // this file is excluded
		}

		funcs, err := findFuncs(file)
		if err != nil {
			return nil, fmt.Errorf("failed parsing funcs from file [%s]: %w", profile.FileName, err)
		}

		comments, err := findComments(file)
		if err != nil { // coverage-ignore
			return nil, fmt.Errorf("failed parsing comments from file [%s]: %w", profile.FileName, err)
		}

		s := Stats{
			Name: noPrefixName,
		}

		for _, f := range funcs {
			c, t := f.coverage(profile, comments)
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

func findComments(filename string) ([]extent, error) {
	fset := token.NewFileSet()

	node, err := parser.ParseFile(fset, filename, nil, parser.ParseComments)
	if err != nil {
		return nil, err //nolint:wrapcheck // relax
	}

	var comments []extent

	for _, c := range node.Comments {
		if strings.Contains(c.Text(), IgnoreText) {
			comments = append(comments, newExtent(fset, c))
		}
	}

	return comments, nil
}

// findFuncs parses the file and returns a slice of FuncExtent descriptors.
func findFuncs(filename string) ([]extent, error) {
	fset := token.NewFileSet()

	parsedFile, err := parser.ParseFile(fset, filename, nil, 0)
	if err != nil {
		return nil, err //nolint:wrapcheck // relax
	}

	visitor := &FuncVisitor{fset: fset}
	ast.Walk(visitor, parsedFile)

	return visitor.funcs, nil
}

// FuncVisitor implements the visitor that builds the function position list for a file.
type FuncVisitor struct {
	fset  *token.FileSet
	funcs []extent
}

// Visit implements the ast.Visitor interface.
func (v *FuncVisitor) Visit(node ast.Node) ast.Visitor {
	if n, ok := node.(*ast.FuncDecl); ok {
		fn := newExtent(v.fset, n)
		v.funcs = append(v.funcs, fn)
	}

	return v
}

type extent struct {
	startLine int
	startCol  int
	endLine   int
	endCol    int
}

func newExtent(fset *token.FileSet, n ast.Node) extent {
	start := fset.Position(n.Pos())
	end := fset.Position(n.End())

	return extent{
		startLine: start.Line,
		startCol:  start.Column,
		endLine:   end.Line,
		endCol:    end.Column,
	}
}

// coverage returns the fraction of the statements in the
// function that were covered, as a numerator and denominator.
func (f extent) coverage(profile *cover.Profile, comments []extent) (int64, int64) {
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

		// add block to coverage statistics only if it was not ignored using comment
		if !hasCommentOnLine(comments, b.StartLine) {
			total += int64(b.NumStmt)

			if b.Count > 0 {
				covered += int64(b.NumStmt)
			}
		}
	}

	return covered, total
}

func hasCommentOnLine(comments []extent, startLine int) bool {
	for _, c := range comments {
		if c.startLine == startLine {
			return true
		}
	}

	return false
}
