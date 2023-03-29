package testcoverage

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strconv"
	"text/tabwriter"
)

func ReportForHuman(result AnalyzeResult, cfg Config) {
	thr := cfg.Threshold

	out := bufio.NewWriter(os.Stdout)
	defer out.Flush()

	{
		fmt.Fprintf(out, "Files meeting coverage threshold of (%d%%):\t", thr.File)
		if len(result.FilesBelowThreshold) > 0 {
			fmt.Fprintf(out, "FAIL")
			report(out, result.FilesBelowThreshold)
		} else {
			fmt.Fprintf(out, "PASS")
		}
	}

	{
		fmt.Fprintf(out, "\nPackages meeting coverage threshold of (%d%%):\t", thr.Package)
		if len(result.PackagesBelowThreshold) > 0 {
			fmt.Fprintf(out, "FAIL")
			report(out, result.PackagesBelowThreshold)
		} else {
			fmt.Fprintf(out, "PASS")
		}
	}

	{
		fmt.Fprintf(out, "\nTotal coverage meeting the threshold of (%d%%):\t", thr.Total)
		if !result.MeetsTotalCoverage {
			fmt.Fprintf(out, "FAIL")
		} else {
			fmt.Fprintf(out, "PASS")
		}
	}

	fmt.Fprintf(out, "\nTotal test coverage: %d%%\n", result.TotalCoverage)
}

func report(w io.Writer, coverageStats []CoverageStats) {
	tabber := tabwriter.NewWriter(w, 1, 8, 1, '\t', 0) //nolint:gomnd // relax
	defer tabber.Flush()

	fmt.Fprintf(tabber, "\n\nIssues with:")

	for _, stats := range coverageStats {
		fmt.Fprintf(tabber, "\n%s\t%d%%", stats.name, stats.CoveredPercentage())
	}

	fmt.Fprintf(tabber, "\n")
}

//nolint:lll // relax
func ReportForGithubAction(result AnalyzeResult, cfg Config) {
	out := bufio.NewWriter(os.Stdout)
	defer out.Flush()

	{
		msg := fmt.Sprintf("::set-output name=total_coverage::%s\n", strconv.Itoa(result.TotalCoverage))
		fmt.Fprint(out, msg)
	}

	for _, stats := range result.FilesBelowThreshold {
		msg := fmt.Sprintf("::error file=%s,line=1::File test coverage below threshold of (%d%%)\n", stats.name, cfg.Threshold.File)
		fmt.Fprint(out, msg)
	}

	for _, stats := range result.PackagesBelowThreshold {
		msg := fmt.Sprintf("::error ::Package (%s) test coverage below threshold of (%d%%)\n", stats.name, cfg.Threshold.Package)
		fmt.Fprint(out, msg)
	}

	if !result.MeetsTotalCoverage {
		msg := fmt.Sprintf("::error ::Total coverage below threshold of (%d%%)\n", cfg.Threshold.Total)
		fmt.Fprint(out, msg)
	}
}
