package testcoverage_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	. "github.com/subhambhardwaj/go-test-coverage/v2/pkg/testcoverage"
)

func TestPackageForFile(t *testing.T) {
	t.Parallel()

	tests := []struct {
		file string
		pkg  string
	}{
		{file: "org.org/project/pkg/foo/bar.go", pkg: "org.org/project/pkg/foo"},
		{file: "pkg/foo/bar.go", pkg: "pkg/foo"},
		{file: "pkg/", pkg: "pkg"},
		{file: "pkg", pkg: "pkg"},
	}

	for _, tc := range tests {
		pkg := PackageForFile(tc.file)
		assert.Equal(t, tc.pkg, pkg)
	}
}
