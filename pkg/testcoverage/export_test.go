package testcoverage

const (
	GaOutputFileEnv       = gaOutputFileEnv
	GaOutputTotalCoverage = gaOutputTotalCoverage
	GaOutputBadgeColor    = gaOutputBadgeColor
	GaOutputBadgeText     = gaOutputBadgeText
	GaOutputReport        = gaOutputReport
)

var (
	MakePackageStats          = makePackageStats
	PackageForFile            = packageForFile
	StoreBadge                = storeBadge
	GenerateAndSaveBadge      = generateAndSaveBadge
	SetOutputValue            = setOutputValue
	LoadBaseCoverageBreakdown = loadBaseCoverageBreakdown
	CompressUncoveredLines    = compressUncoveredLines
	ReportUncoveredLines      = reportUncoveredLines
	StatusStr                 = statusStr
)

type (
	StorerFactories = storerFactories
)
