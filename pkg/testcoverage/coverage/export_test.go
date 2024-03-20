package coverage

var (
	FindFile        = findFile
	FindComments    = findComments
	FindFuncs       = findFuncs
	ParseProfiles   = parseProfiles
	CoverageForFile = coverageForFile
)

type Extent = extent
