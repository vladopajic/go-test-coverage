package coverage

var (
	FindFile           = findFileCreator()
	FindAnnotations    = findAnnotations
	FindFuncsAndBlocks = findFuncsAndBlocks
	ParseProfiles      = parseProfiles
	SumCoverage        = sumCoverage
)

type Extent = extent
