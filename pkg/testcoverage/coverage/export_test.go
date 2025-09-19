package coverage

var (
	FindFileCreator    = findFileCreator
	FindAnnotations    = findAnnotations
	FindFuncsAndBlocks = findFuncsAndBlocks
	ParseProfiles      = parseProfiles
	SumCoverage        = sumCoverage
	FindGoModFile      = findGoModFile
)

type Extent = extent
