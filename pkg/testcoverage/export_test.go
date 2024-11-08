package testcoverage

const (
	GaOutputFileEnv       = gaOutputFileEnv
	GaOutputTotalCoverage = gaOutputTotalCoverage
	GaOutputBadgeColor    = gaOutputBadgeColor
	GaOutputBadgeText     = gaOutputBadgeText
	GaOutputReport        = gaOutputReport
)

var (
	MakePackageStats     = makePackageStats
	PackageForFile       = packageForFile
	StoreBadge           = storeBadge
	GenerateAndSaveBadge = generateAndSaveBadge
	SetOutputValue       = setOutputValue
)

type (
	StorerFactories = storerFactories
)
