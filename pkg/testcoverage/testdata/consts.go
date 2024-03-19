package testdata

const (
	// this is valid profile with valid data.
	// it is made at earlier point in time so it does not need to reflect
	// the most recent profile
	ProfileOK = "ok.profile"

	// this profile is syntetically made with full coverage
	ProfileOKFull = "ok_full.profile"

	// just like `ok.profile` but does not have entires for `path/path.go` file
	ProfileOKNoPath = "ok_no_path.profile"

	// contains profile item with invalid file
	ProfileNOK = "nok.profile"

	// contains profile items for `path/path.go` file, but
	// does not have all profile items
	ProfileNOKInvalidLength = "invalid_length.profile"

	// contains profile items for `path/path.go` file, but
	// does not have correct profile items
	ProfileNOKInvalidData = "invalid_data.profile"
)
