package coverage

import (
	"fmt"
	"go/ast"
	"go/build"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"golang.org/x/tools/cover"

	"github.com/vladopajic/go-test-coverage/v2/pkg/testcoverage/path"
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

	files, err := findFiles(profiles, cfg.LocalPrefix)
	if err != nil {
		return nil, err
	}

	fileStats := make([]Stats, 0, len(profiles))
	excludeRules := compileExcludePathRules(cfg.ExcludePaths)

	for _, profile := range profiles {
		fi, ok := files[profile.FileName]
		if !ok { // coverage-ignore
			// should already be handled above, but let's check it again
			return nil, fmt.Errorf("could not find file [%s]", profile.FileName)
		}

		if ok := matches(excludeRules, fi.noPrefixName); ok {
			continue // this file is excluded
		}

		s, err := coverageForFile(profile, fi)
		if err != nil {
			return nil, err
		}

		if s.Total == 0 {
			// do not include files that doesn't have statements.
			// this can happen when everything is excluded with comment annotations, or
			// simply file doesn't have any statement.
			//
			// note: we are explicitly adding `continue` statement, instead of having code like this:
			// if s.Total != 0 {
			// 	fileStats = append(fileStats, s)
			// }
			// because with `continue` add additional statements in coverage profile which will require
			// to have it covered with tests. since this is interesting case, to have it covered
			// with tests, we have code written in this way
			continue
		}

		fileStats = append(fileStats, s)
	}

	return fileStats, nil
}

func coverageForFile(profile *cover.Profile, fi fileInfo) (Stats, error) {
	source, err := os.ReadFile(fi.path)
	if err != nil { // coverage-ignore
		return Stats{}, fmt.Errorf("failed reading file source [%s]: %w", fi.path, err)
	}

	funcs, blocks, err := findFuncsAndBlocks(source)
	if err != nil { // coverage-ignore
		return Stats{}, err
	}

	annotations, err := findAnnotations(source)
	if err != nil { // coverage-ignore
		return Stats{}, err
	}

	s := sumCoverage(profile, funcs, blocks, annotations)
	s.Name = fi.noPrefixName

	return s, nil
}

type fileInfo struct {
	path         string
	noPrefixName string
}

func findFiles(profiles []*cover.Profile, prefix string) (map[string]fileInfo, error) {
	result := make(map[string]fileInfo)
	findFile := findFileCreator()

	for _, profile := range profiles {
		file, noPrefixName, found := findFile(profile.FileName, prefix)
		if !found {
			return nil, fmt.Errorf("could not find file [%s]", profile.FileName)
		}

		result[profile.FileName] = fileInfo{
			path:         file,
			noPrefixName: noPrefixName,
		}
	}

	return result, nil
}

func findFileCreator() func(file, prefix string) (string, string, bool) {
	cache := make(map[string]*build.Package)
	files := []string(nil)

	findRelative := func(file, prefix string) (string, string, bool) {
		noPrefixName := stripPrefix(file, prefix)
		_, err := os.Stat(noPrefixName)

		return noPrefixName, noPrefixName, err == nil
	}

	findBuildImport := func(file string) (string, string, bool) {
		dir, file := filepath.Split(file)
		pkg, exists := cache[dir]

		if !exists {
			var err error

			pkg, err = build.Import(dir, ".", build.FindOnly)
			if err != nil {
				return "", "", false
			}

			cache[dir] = pkg
		}

		file = filepath.Join(pkg.Dir, file)
		_, err := os.Stat(file)
		noPrefixName := stripPrefix(path.NormalizeForTool(file), path.NormalizeForTool(pkg.Root))

		return file, noPrefixName, err == nil
	}

	findWalk := func(file, prefix string) (string, string, bool) {
		if files == nil {
			files = listAllFiles("./")
		}

		noPrefixName := stripPrefix(file, prefix)
		f, found := hasFile(files, noPrefixName)

		return path.NormalizeForOS(f), noPrefixName, found
	}

	return func(file, prefix string) (string, string, bool) {
		if fPath, fNoPrefix, found := findRelative(file, prefix); found { // coverage-ignore
			return fPath, fNoPrefix, found
		}

		if fPath, fNoPrefix, found := findWalk(file, prefix); found {
			return fPath, fNoPrefix, found
		}

		if fPath, fNoPrefix, found := findBuildImport(file); found {
			return fPath, fNoPrefix, found
		}

		return "", "", false
	}
}

func listAllFiles(rootDir string) []string {
	var files []string

	err := filepath.Walk(rootDir, func(file string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !info.IsDir() &&
			strings.HasSuffix(file, ".go") &&
			!strings.HasSuffix(file, "_test.go") {
			files = append(files, path.NormalizeForTool(file))
		}

		return nil
	})

	if err != nil {
		panic(err)
		// return nil
	}

	return files
}

func hasFile(files []string, search string) (string, bool) {
	var result string

	for _, f := range files {
		if strings.HasSuffix(f, search) {
			if result != "" {
				return "", false
			}

			result = f
		}
	}

	return result, result != ""
}

func findAnnotations(source []byte) ([]extent, error) {
	fset := token.NewFileSet()

	node, err := parser.ParseFile(fset, "", source, parser.ParseComments)
	if err != nil {
		return nil, fmt.Errorf("can't parse comments: %w", err)
	}

	var res []extent

	for _, c := range node.Comments {
		if strings.Contains(c.Text(), IgnoreText) {
			res = append(res, newExtent(fset, c))
		}
	}

	return res, nil
}

func findFuncsAndBlocks(source []byte) ([]extent, []extent, error) {
	fset := token.NewFileSet()

	parsedFile, err := parser.ParseFile(fset, "", source, 0)
	if err != nil {
		return nil, nil, fmt.Errorf("can't parse source: %w", err)
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
		v.funcs = append(v.funcs, newExtent(v.fset, n.Body))

	case *ast.IfStmt:
		v.addBlock(n.Body)
	case *ast.SwitchStmt:
		v.addBlock(n.Body)
	case *ast.TypeSwitchStmt:
		v.addBlock(n.Body)
	case *ast.SelectStmt: // coverage-ignore
		v.addBlock(n.Body)
	case *ast.ForStmt:
		v.addBlock(n.Body)
	case *ast.RangeStmt:
		v.addBlock(n.Body)
	}

	return v
}

func (v *visitor) addBlock(n ast.Node) {
	v.blocks = append(v.blocks, newExtent(v.fset, n))
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

func sumCoverage(profile *cover.Profile, funcs, blocks, annotations []extent) Stats {
	s := Stats{}

	for _, f := range funcs {
		c, t, ul := coverage(profile, f, blocks, annotations)
		s.Total += t
		s.Covered += c
		s.UncoveredLines = append(s.UncoveredLines, ul...)
	}

	s.UncoveredLines = dedup(s.UncoveredLines)

	return s
}

// coverage returns the fraction of the statements in the
// function that were covered, as a numerator and denominator.
//
//nolint:cyclop,gocognit,maintidx // relax
func coverage(
	profile *cover.Profile,
	f extent,
	blocks, annotations []extent,
) (int64, int64, []int) {
	if hasExtentWithStartLine(annotations, f.StartLine) {
		// case when entire function is ignored
		return 0, 0, nil
	}

	var (
		covered, total int64
		skip           extent
		uncoveredLines []int
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

		// add block to coverage statistics only if it was not ignored using comment annotations
		if hasExtentWithStartLine(annotations, b.StartLine) {
			if e, found := findExtentWithStartLine(blocks, b.StartLine); found {
				skip = e
			}

			continue
		}

		total += int64(b.NumStmt)

		if b.Count > 0 {
			covered += int64(b.NumStmt)
		} else {
			for i := range (b.EndLine - b.StartLine) + 1 {
				uncoveredLines = append(uncoveredLines, b.StartLine+i)
			}
		}
	}

	return covered, total, uncoveredLines
}

func dedup(ss []int) []int {
	if len(ss) <= 1 {
		return ss
	}

	sort.Ints(ss)
	result := []int{ss[0]}

	for i := 1; i < len(ss); i++ {
		if ss[i] != ss[i-1] {
			result = append(result, ss[i])
		}
	}

	return result
}
