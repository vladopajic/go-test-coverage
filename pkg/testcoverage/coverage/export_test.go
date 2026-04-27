package coverage

var (
	FindFileCreator             = findFileCreator
	FindAnnotations             = findAnnotations
	FindFuncsAndBlocks          = findFuncsAndBlocks
	ParseProfiles               = parseProfiles
	SumCoverage                 = sumCoverage
	FindGoModFile               = findGoModFile
	PluckStartLine              = pluckStartLine
	FindFilePathMatchingSearch = findFilePathMatchingSearch
)

type Extent = extent
type FileInfo = fileInfo

func NewFileInfo(name string) fileInfo {
	return fileInfo{name: name}
}
