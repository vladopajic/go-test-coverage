package testcoverage

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"
	"text/tabwriter"
)

//nolint:wsl // relax
func Analyze(cfg Config, coverageStats []CoverageStats) bool {
	thr := cfg.Threshold

	out := bufio.NewWriter(os.Stdout)
	defer out.Flush()

	filesBelowThreshold := checkCoverageStatsBelowThreshold(coverageStats, thr.File)
	packagesBelowThreshold := checkCoverageStatsBelowThreshold(
		makePackageStats(coverageStats), thr.Package,
	)
	totalStats := calcTotalStats(coverageStats)
	meetsTotalCoverage := totalStats.CoveredPercentage() >= thr.Total

	fmt.Fprintf(out, "Files test coverage meeting the threshold\t(%d%%): ", thr.File)
	if len(filesBelowThreshold) > 0 {
		fmt.Fprintf(out, "FAIL")
		report(out, filesBelowThreshold, cfg.LocalPrefix)
	} else {
		fmt.Fprintf(out, "PASS")
	}

	fmt.Fprintf(out, "\nPackages test coverage meeting the threshold\t(%d%%): ", thr.Package)
	if len(packagesBelowThreshold) > 0 {
		fmt.Fprintf(out, "FAIL")
		report(out, packagesBelowThreshold, cfg.LocalPrefix)
	} else {
		fmt.Fprintf(out, "PASS")
	}

	fmt.Fprintf(out, "\nTotal test coverage meeting the threshold\t(%d%%): ", thr.Total)
	if !meetsTotalCoverage {
		fmt.Fprintf(out, "FAIL")
	} else {
		fmt.Fprintf(out, "PASS")
	}

	fmt.Fprintf(out, "\nTotal test coverage: %d%%\n", totalStats.CoveredPercentage())

	return len(filesBelowThreshold) == 0 && len(packagesBelowThreshold) == 0 && meetsTotalCoverage
}

func report(w io.Writer, coverageStats []CoverageStats, localPrefix string) {
	localPrefix += "/"

	tabber := tabwriter.NewWriter(w, 1, 8, 1, '\t', 0) //nolint:gomnd // relax
	defer tabber.Flush()

	for _, stats := range coverageStats {
		name := strings.Replace(stats.name, localPrefix, "", 1)
		fmt.Fprintf(tabber, "\n%s\t%d%%", name, stats.CoveredPercentage())
	}

	fmt.Fprintf(tabber, "\n")
}
