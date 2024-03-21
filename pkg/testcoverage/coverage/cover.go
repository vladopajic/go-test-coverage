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

		funcs, blocks, err := findFuncsAndBlocks(source)
		if err != nil { // coverage-ignore
			return nil, err
		}

		comments, err := findComments(source)
		if err != nil { // coverage-ignore
			return nil, err
		}

		s := coverageForFile(profile, funcs, blocks, comments)
		if s.Total == 0 {
			// do not include files that doesn't have statements
			// everything could be excluded with comments or simply file doesn't have them
			continue
		}

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

func findFuncsAndBlocks(source []byte) ([]extent, []extent, error) {
	fset := token.NewFileSet()

	parsedFile, err := parser.ParseFile(fset, "", source, 0)
	if err != nil {
		return nil, nil, err //nolint:wrapcheck // relax
	}

	v := &visitor{fset: fset}
	ast.Walk(v, parsedFile)

	return v.funcs, v.blocks, nil
}

type visitor struct {
	fset   *token.FileSet
	funcs  []extent
	blocks []extent
}

// Visit implements the ast.Visitor interface.
func (v *visitor) Visit(node ast.Node) ast.Visitor {
	switch n := node.(type) {
	case *ast.FuncDecl:
		fn := newExtent(v.fset, n.Body)
		v.funcs = append(v.funcs, fn)

	case *ast.IfStmt:
		fn := newExtent(v.fset, n.Body)
		v.blocks = append(v.blocks, fn)
	case *ast.ForStmt:
		fn := newExtent(v.fset, n.Body)
		v.blocks = append(v.blocks, fn)
	case *ast.RangeStmt:
		fn := newExtent(v.fset, n.Body)
		v.blocks = append(v.blocks, fn)
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
//nolint:cyclop,gocognit // relax
func (f extent) coverage(profile *cover.Profile, blocks, comments []extent) (int64, int64) {
	if hasExtentWithStartLine(comments, f.StartLine) {
		// case when entire function is ignored
		return 0, 0
	}

	var (
		covered, total int64
		skip           extent
	)

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

		if b.StartLine < skip.EndLine || (b.EndLine == f.StartLine && b.StartCol <= skip.EndCol) {
			// this block has comment annotation
			continue
		}

		// add block to coverage statistics only if it was not ignored using comment
		if hasExtentWithStartLine(comments, b.StartLine) {
			if e, found := findExtentWithStartLine(blocks, b.StartLine); found {
				skip = e
			}

			continue
		}

		total += int64(b.NumStmt)

		if b.Count > 0 {
			covered += int64(b.NumStmt)
		}
	}

	return covered, total
}

func findExtentWithStartLine(ee []extent, line int) (extent, bool) {
	for _, e := range ee {
		if e.StartLine <= line && e.EndLine >= line {
			return e, true
		}
	}

	return extent{}, false
}

func hasExtentWithStartLine(ee []extent, startLine int) bool {
	_, found := findExtentWithStartLine(ee, startLine)
	return found
}

func coverageForFile(profile *cover.Profile, funcs, blocks, comments []extent) Stats {
	s := Stats{}

	for _, f := range funcs {
		c, t := f.coverage(profile, blocks, comments)
		s.Total += t
		s.Covered += c
	}

	return s
}
