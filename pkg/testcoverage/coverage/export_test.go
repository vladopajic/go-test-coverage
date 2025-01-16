package coverage

var (
	FindFile           = findFileCreator()
	FindAnnotations    = findAnnotations
	FindFuncsAndBlocks = findFuncsAndBlocks
	ParseProfiles      = parseProfiles
	CoverageForFile    = coverageForFile
)

type Extent = extent
