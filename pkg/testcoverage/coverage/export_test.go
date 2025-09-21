package coverage

var (
	FindFileCreator    = findFileCreator
	FindAnnotations    = findAnnotations
	FindFuncsAndBlocks = findFuncsAndBlocks
	ParseProfiles      = parseProfiles
	SumCoverage        = sumCoverage
	FindGoModFile      = findGoModFile
	PluckStartLine     = pluckStartLine
)

type Extent = extent
