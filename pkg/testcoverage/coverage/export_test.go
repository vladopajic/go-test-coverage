package coverage

var (
	FindFile           = findFile
	FindComments       = findComments
	FindFuncsAndBlocks = findFuncsAndBlocks
	ParseProfiles      = parseProfiles
	CoverageForFile    = coverageForFile
)

type Extent = extent
