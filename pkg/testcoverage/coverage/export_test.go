package coverage

var (
	FindFileCreator    = findFileCreator
	FindAnnotations    = findAnnotations
	FindFuncsAndBlocks = findFuncsAndBlocks
	ParseProfiles      = parseProfiles
	SumCoverage        = sumCoverage
)

type Extent = extent
