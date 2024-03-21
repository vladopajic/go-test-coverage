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
	Profiles     []string
	LocalPrefix  string
	ExcludePaths []string
}

func GenerateCoverageStats(cfg Config) ([]Stats, error) {
	profiles, err := parseProfiles(cfg.Profiles)
	if err != nil {
		return nil, fmt.Errorf("parsing profiles: %w", err)
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

		source, err := readFileSource(file)
		if err != nil { // coverage-ignore
			return nil, fmt.Errorf("failed reading file source [%s]: %w", profile.FileName, err)
		}

		funcs, err := findFuncs(source)
		if err != nil { // coverage-ignore
			return nil, err
		}

		comments, err := findComments(source)
		if err != nil { // coverage-ignore
			return nil, err
		}

		s := coverageForFile(profile, funcs, comments)
		s.Name = noPrefixName
		fileStats = append(fileStats, s)
	}

	return fileStats, nil
}

// findFile finds the location of the named file in GOROOT, GOPATH etc.
//
//nolint:goerr113 // relax
func findFile(file, prefix string) (string, string, error) {
	profileFile := file

	noPrefixName := stripPrefix(file, prefix)
	if _, err := os.Stat(noPrefixName); err == nil { // coverage-ignore
		return noPrefixName, noPrefixName, nil
	}

	dir, file := filepath.Split(file)

	pkg, err := build.Import(dir, ".", build.FindOnly)
	if err != nil {
		return "", "", fmt.Errorf("can't find file %q: %w", profileFile, err)
	}

	file = filepath.Join(pkg.Dir, file)
	if _, err := os.Stat(file); err == nil {
		return file, stripPrefix(file, pkg.Root), nil
	}

	return "", "", fmt.Errorf("can't find file %q", profileFile)
}

func readFileSource(filename string) ([]byte, error) {
	return os.ReadFile(filename) //nolint:wrapcheck // relax
}

func findComments(source []byte) ([]extent, error) {
	fset := token.NewFileSet()

	node, err := parser.ParseFile(fset, "", source, parser.ParseComments)
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
func findFuncs(source []byte) ([]extent, error) {
	fset := token.NewFileSet()

	parsedFile, err := parser.ParseFile(fset, "", source, 0)
	if err != nil {
		return nil, err //nolint:wrapcheck // relax
	}

	visitor := &funcVisitor{fset: fset}
	ast.Walk(visitor, parsedFile)

	return visitor.funcs, nil
}

// funcVisitor implements the visitor that builds the function position list for a file.
type funcVisitor struct {
	fset  *token.FileSet
	funcs []extent
}

// Visit implements the ast.Visitor interface.
func (v *funcVisitor) Visit(node ast.Node) ast.Visitor {
	if n, ok := node.(*ast.FuncDecl); ok {
		fn := newExtent(v.fset, n)
		v.funcs = append(v.funcs, fn)
	}

	return v
}

type extent struct {
	StartLine int
	StartCol  int
	EndLine   int
	EndCol    int
}

func newExtent(fset *token.FileSet, n ast.Node) extent {
	start := fset.Position(n.Pos())
	end := fset.Position(n.End())

	return extent{
		StartLine: start.Line,
		StartCol:  start.Column,
		EndLine:   end.Line,
		EndCol:    end.Column,
	}
}

// coverage returns the fraction of the statements in the
// function that were covered, as a numerator and denominator.
//
//nolint:cyclop // relax
func (f extent) coverage(profile *cover.Profile, comments []extent) (int64, int64) {
	if hasCommentOnLine(comments, f.StartLine) {
		// case when entire function is ignored
		return 0, 0
	}

	var covered, total int64

	// the blocks are sorted, so we can stop counting as soon as
	// we reach the end of the relevant block.
	for _, b := range profile.Blocks {
		if b.StartLine > f.EndLine || (b.StartLine == f.EndLine && b.StartCol >= f.EndCol) {
			// past the end of the function.
			break
		}

		if b.EndLine < f.StartLine || (b.EndLine == f.StartLine && b.EndCol <= f.StartCol) {
			// before the beginning of the function
			continue
		}

		if hasCommentOnLine(comments, b.StartLine) {
			// add block to coverage statistics only if it was not ignored using comment
			continue
		}

		total += int64(b.NumStmt)

		if b.Count > 0 {
			covered += int64(b.NumStmt)
		}
	}

	return covered, total
}

func hasCommentOnLine(comments []extent, startLine int) bool {
	for _, c := range comments {
		if c.StartLine == startLine {
			return true
		}
	}

	return false
}

func coverageForFile(profile *cover.Profile, funcs, comments []extent) Stats {
	s := Stats{}

	for _, f := range funcs {
		c, t := f.coverage(profile, comments)
		s.Total += t
		s.Covered += c
	}

	return s
}
