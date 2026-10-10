package coverage

import "sync"

var (
	FindFileCreator            = findFileCreator
	FindAnnotations            = findAnnotations
	FindFuncsAndBlocks         = findFuncsAndBlocks
	FindTrivialErrorChecks     = findTrivialErrorChecks
	ParseProfiles              = parseProfiles
	SumCoverage                = sumCoverage
	FindGoModFile              = findGoModFile
	PluckStartLine             = pluckStartLine
	FindFilePathMatchingSearch = findFilePathMatchingSearch
)

type (
	Extent   = extent
	FileInfo = fileInfo
)

func NewFileInfo(name string) fileInfo {
	return fileInfo{name: name, path: name}
}

func ResetPackagesCache() {
	pkgs = nil
	pkgsOnce = sync.Once{}
}
