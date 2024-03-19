package coverage

import (
	"fmt"
	"slices"
	"strings"

	"golang.org/x/tools/cover"
)

func parseProfiles(paths []string) ([]*cover.Profile, error) {
	var result []*cover.Profile

	for _, path := range paths {
		profiles, err := cover.ParseProfiles(path)
		if err != nil {
			return nil, fmt.Errorf("parsing profile file: %w", err)
		}

		if result == nil {
			result = profiles
			continue
		}

		result, err = mergeProfiles(result, profiles)
		if err != nil {
			return nil, fmt.Errorf("merging profiles: %w", err)
		}
	}

	slices.SortFunc(result, func(a, b *cover.Profile) int {
		return strings.Compare(a.FileName, b.FileName)
	})

	return result, nil
}

func mergeProfiles(a, b []*cover.Profile) ([]*cover.Profile, error) {
	for _, pb := range b {
		if idx, found := findProfileForFile(a, pb.FileName); found {
			m, err := mergeSameFileProfile(a[idx], pb)
			if err != nil {
				return nil, err
			}

			a[idx] = m
		} else {
			a = append(a, pb)
		}
	}

	return a, nil
}

func findProfileForFile(profiles []*cover.Profile, file string) (int, bool) {
	for i, p := range profiles {
		if p.FileName == file {
			return i, true
		}
	}

	return -1, false
}

//nolint:goerr113 // relax
func mergeSameFileProfile(ap, bp *cover.Profile) (*cover.Profile, error) {
	if len(ap.Blocks) != len(bp.Blocks) {
		return nil, fmt.Errorf("inconsistent profiles length [%q, %q]", ap.FileName, bp.FileName)
	}

	for i := 0; i < len(ap.Blocks); i++ {
		a, b := ap.Blocks[i], bp.Blocks[i]

		if b.StartLine == a.StartLine &&
			b.StartCol == a.StartCol &&
			b.EndLine == a.EndLine &&
			b.EndCol == a.EndCol &&
			b.NumStmt == a.NumStmt {
			ap.Blocks[i].Count = max(a.Count, b.Count)
		} else {
			return nil, fmt.Errorf("inconsistent profile data [%q, %q]", ap.FileName, bp.FileName)
		}
	}

	return ap, nil
}
