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
		fmt.Fprintf(tabber, "\n%s\t%d%%", stats.Name, stats.CoveredPercentage())
	}

	fmt.Fprintf(tabber, "\n")
}

func ReportForGithubAction(w io.Writer, result AnalyzeResult, cfg Config) {
	out := bufio.NewWriter(w)
	defer out.Flush()

	reportLineError := func(file, title, msg string) {
		fmt.Fprintf(out, "::error file=%s,title=%s,line=1::%s\n", file, title, msg)
	}
	reportError := func(title, msg string) {
		fmt.Fprintf(out, "::error title=%s::%s\n", title, msg)
	}

	for _, stats := range result.FilesBelowThreshold {
		title := "File test coverage below threshold"
		c := stats.CoveredPercentage()
		t := cfg.Threshold.File
		msg := fmt.Sprintf("coverage: %d%%; threshold: %d%%", c, t)
		reportLineError(stats.Name, title, msg)
	}

	for _, stats := range result.PackagesBelowThreshold {
		title := "Package test coverage below threshold"
		c := stats.CoveredPercentage()
		t := cfg.Threshold.Package
		msg := fmt.Sprintf("package: %s; coverage: %d%%; threshold: %d%%", stats.Name, c, t)
		reportError(title, msg)
	}

	if !result.MeetsTotalCoverage {
		title := "Total test coverage below threshold"
		c := result.TotalCoverage
		t := cfg.Threshold.Total
		msg := fmt.Sprintf("coverage: %d%%; threshold: %d%%", c, t)
		reportError(title, msg)
	}
}

func SetGithubActionOutput(result AnalyzeResult) error {
	githubOutputFile, err := openGitHubOutput(os.Getenv("GITHUB_OUTPUT"))
	if err != nil {
		return fmt.Errorf("could not open GITHUB_OUTPUT file: %w", err)
	}
	defer githubOutputFile.Close()

	err = setOutput(githubOutputFile, "total_coverage", strconv.Itoa(result.TotalCoverage))
	if err != nil {
		return fmt.Errorf("failed setting github output: %w", err)
	}

	return nil
}

func openGitHubOutput(p string) (io.WriteCloser, error) {
	//nolint:gomnd,wrapcheck //relax
	return os.OpenFile(p, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
}

func setOutput(w io.Writer, name, value string) error {
	if _, err := w.Write([]byte(fmt.Sprintf("%s=%s\n", name, value))); err != nil {
		return fmt.Errorf("failed write: %w", err)
	}

	return nil
}
