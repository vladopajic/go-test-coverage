package testcoverage

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func pushD(t *testing.T, wd string) func() {
	oldWd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}

	err = os.Chdir(wd)
	if err != nil {
		t.Fatal(err)
	}

	popD := func() {
		err = os.Chdir(oldWd)
		if err != nil {
			t.Fatal(err)
		}
	}

	return popD
}

func Test_100PercentSuppression(t *testing.T) {
	popD := pushD(t, "testdata/functional_test")
	defer popD()
	var err error
	outfile, err := filepath.Abs(filepath.Join(t.TempDir(), "functional_test.coverage"))
	if err != nil {
		t.Fatal(err)
	}

	cmd := exec.Command("go", "test", "-cover", "-tags", "sample", "-coverprofile="+outfile, "./...")

	print("command: " + cmd.String())
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatal(err)
	}

	data, err := os.ReadFile(outfile)
	if err != nil {
		t.Fatal(err)
	}

	println(string(out))
	println("data: " + string(data))

	cfg := Config{
		Profile:         outfile,
		ReportUncovered: true,
	}
	collector := &bytes.Buffer{}
	ok := Check(collector, cfg)
	if !ok {
		t.Fatal("check failed; console output is...\n " + collector.String())
	}

	require.Contains(t, collector.String(), "Total test coverage: 100%")
}
