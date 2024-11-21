package testdata

const (
	// this is valid profile with valid data.
	// it is made at earlier point in time so it does not need to reflect
	// the most recent profile
	ProfileOK = "ok.profile"

	// this profile is synthetically made with full coverage
	ProfileOKFull = "ok_full.profile"

	// just like `ok.profile` but does not have entries for `badge/generate.go` file
	ProfileOKNoBadge = "ok_no_badge.profile"

	// this profile has no statements for file
	ProfileOKNoStatements = "ok_no_statements.profile"

	// contains profile item with invalid file
	ProfileNOK = "nok.profile"

	// contains profile items for `badge/generate.go` file, but
	// does not have all profile items
	ProfileNOKInvalidLength = "invalid_length.profile"

	// contains profile items for `badge/generate.go` file, but
	// does not have correct profile items
	ProfileNOKInvalidData = "invalid_data.profile"

	// holds valid test coverage breakdown
	BreakdownOK = "breakdown_ok.testcoverage"

	// holds invalid test coverage breakdown
	BreakdownNOK = "breakdown_nok.testcoverage"
)
