package testcoverage

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
	"text/tabwriter"

	"github.com/vladopajic/go-test-coverage/v2/pkg/testcoverage/badge"
	"github.com/vladopajic/go-test-coverage/v2/pkg/testcoverage/coverage"
)

func ReportForHuman(w io.Writer, result AnalyzeResult) {
	out := bufio.NewWriter(w)
	defer out.Flush()

	tabber := tabwriter.NewWriter(out, 1, 8, 2, '\t', 0) //nolint:mnd // relax
	defer tabber.Flush()

	statusStr := func(passing bool) string {
		if passing {
			return "PASS"
		}

		return "FAIL"
	}

	thr := result.Threshold

	// File threshold report
	fmt.Fprintf(tabber, "File coverage threshold (%d%%) satisfied:\t", thr.File)
	fmt.Fprint(tabber, statusStr(len(result.FilesBelowThreshold) == 0))
	reportIssuesForHuman(tabber, result.FilesBelowThreshold)

	// Package threshold report
	fmt.Fprintf(tabber, "\nPackage coverage threshold (%d%%) satisfied:\t", thr.Package)
	fmt.Fprint(tabber, statusStr(len(result.PackagesBelowThreshold) == 0))
	reportIssuesForHuman(tabber, result.PackagesBelowThreshold)

	// Total threshold report
	fmt.Fprintf(tabber, "\nTotal coverage threshold (%d%%) satisfied:\t", thr.Total)
	fmt.Fprint(tabber, statusStr(result.MeetsTotalCoverage))

	fmt.Fprintf(tabber, "\nTotal test coverage: %d%%\n", result.TotalCoverage)
}

func reportIssuesForHuman(w io.Writer, coverageStats []coverage.Stats) {
	if len(coverageStats) == 0 {
		return
	}

	fmt.Fprintf(w, "\n  below threshold:\tcoverage:\tthreshold:")

	for _, stats := range coverageStats {
		fmt.Fprintf(w, "\n  %s\t%d%%\t%d%%", stats.Name, stats.CoveredPercentage(), stats.Threshold)
	}

	fmt.Fprintf(w, "\n")
}

func ReportForGithubAction(w io.Writer, result AnalyzeResult) {
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
		msg := fmt.Sprintf(
			"%s: coverage: %d%%; threshold: %d%%",
			title, stats.CoveredPercentage(), stats.Threshold,
		)
		reportLineError(stats.Name, title, msg)
	}

	for _, stats := range result.PackagesBelowThreshold {
		title := "Package test coverage below threshold"
		msg := fmt.Sprintf(
			"%s: package: %s; coverage: %d%%; threshold: %d%%",
			title, stats.Name, stats.CoveredPercentage(), stats.Threshold,
		)
		reportError(title, msg)
	}

	if !result.MeetsTotalCoverage {
		title := "Total test coverage below threshold"
		msg := fmt.Sprintf(
			"%s: coverage: %d%%; threshold: %d%%",
			title, result.TotalCoverage, result.Threshold.Total,
		)
		reportError(title, msg)
	}
}

const (
	gaOutputFileEnv       = "GITHUB_OUTPUT"
	gaOutputTotalCoverage = "total-coverage"
	gaOutputBadgeColor    = "badge-color"
	gaOutputBadgeText     = "badge-text"
	gaOutputReport        = "report"
)

func SetGithubActionOutput(result AnalyzeResult, report string) error {
	file, err := openGitHubOutput(os.Getenv(gaOutputFileEnv))
	if err != nil {
		return fmt.Errorf("could not open GitHub output file: %w", err)
	}

	totalStr := strconv.Itoa(result.TotalCoverage)

	return errors.Join(
		setOutputValue(file, gaOutputTotalCoverage, totalStr),
		setOutputValue(file, gaOutputBadgeColor, badge.Color(result.TotalCoverage)),
		setOutputValue(file, gaOutputBadgeText, totalStr+"%"),
		setOutputValue(file, gaOutputReport, multiline(report)),
		file.Close(),
	)
}

func openGitHubOutput(p string) (io.WriteCloser, error) {
	//nolint:mnd,wrapcheck // error is wrapped at level above
	return os.OpenFile(p, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
}

func setOutputValue(w io.Writer, name, value string) error {
	data := []byte(fmt.Sprintf("%s=%s\n", name, value))

	_, err := w.Write(data)
	if err != nil {
		return fmt.Errorf("set output for [%s]: %w", name, err)
	}

	return nil
}

func multiline(s string) string {
	s = strings.ReplaceAll(s, "\n", "%0A")
	s = strings.ReplaceAll(s, "\r", "%0D")
	s = strings.ReplaceAll(s, "%", "%25")

	return s
}
