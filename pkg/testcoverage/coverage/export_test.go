package coverage

var (
	FindFile           = findFile
	FindAnnotations    = findAnnotations
	FindFuncsAndBlocks = findFuncsAndBlocks
	ParseProfiles      = parseProfiles
	CoverageForFile    = coverageForFile
)

type Extent = extent
