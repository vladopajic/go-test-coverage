package testcoverage

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"strconv"
	"text/tabwriter"

	"github.com/vladopajic/go-test-coverage/v2/pkg/testcoverage/badge"
	"github.com/vladopajic/go-test-coverage/v2/pkg/testcoverage/coverage"
)

func ReportForHuman(w io.Writer, result AnalyzeResult) {
	out := bufio.NewWriter(w)
	defer out.Flush()

	reportCoverage(out, result)
	reportUncoveredLines(out, result)
	reportMissingExplanations(out, result)
	reportDiff(out, result)
}

func reportCoverage(w io.Writer, result AnalyzeResult) {
	tabber := tabwriter.NewWriter(w, 1, 8, 2, '\t', 0) //nolint:mnd // relax
	defer tabber.Flush()

	thr := result.Threshold

	if thr.File > 0 || result.HasFileOverrides { // File threshold report
		fmt.Fprintf(tabber, "File coverage threshold (%d%%) satisfied:\t", thr.File)
		fmt.Fprint(tabber, statusStr(len(result.FilesBelowThreshold) == 0))
		reportIssuesForHuman(tabber, result.FilesBelowThreshold)
		fmt.Fprint(tabber, "\n")
	}

	if thr.Package > 0 || result.HasPackageOverrides { // Package threshold report
		fmt.Fprintf(tabber, "Package coverage threshold (%d%%) satisfied:\t", thr.Package)
		fmt.Fprint(tabber, statusStr(len(result.PackagesBelowThreshold) == 0))
		reportIssuesForHuman(tabber, result.PackagesBelowThreshold)
		fmt.Fprint(tabber, "\n")
	}

	if thr.Total > 0 { // Total threshold report
		fmt.Fprintf(tabber, "Total coverage threshold (%d%%) satisfied:\t", thr.Total)
		fmt.Fprint(tabber, statusStr(result.MeetsTotalCoverage()))
		fmt.Fprint(tabber, "\n")
	}

	fmt.Fprintf(tabber, "Total test coverage: %s\n", result.TotalStats.Str())
}

func reportIssuesForHuman(w io.Writer, coverageStats []coverage.Stats) {
	if len(coverageStats) == 0 {
		return
	}

	fmt.Fprintf(w, "\n  below threshold:\tcoverage:\tthreshold:")

	for _, stats := range coverageStats {
		fmt.Fprintf(w, "\n  %s\t%s\t%d%%", stats.Name, stats.Str(), stats.Threshold)
	}

	fmt.Fprintf(w, "\n")
}

func reportUncoveredLines(w io.Writer, result AnalyzeResult) {
	if result.Pass() || len(result.FilesWithUncoveredLines) == 0 {
		return
	}

	tabber := tabwriter.NewWriter(w, 1, 8, 2, '\t', 0) //nolint:mnd // relax
	defer tabber.Flush()

	fmt.Fprintf(tabber, "\nFiles with uncovered lines:")
	fmt.Fprintf(tabber, "\n  file:\tuncovered lines:")

	for _, stats := range result.FilesWithUncoveredLines {
		if len(stats.UncoveredLines) > 0 {
			fmt.Fprintf(tabber, "\n  %s\t", stats.Name)
			compressUncoveredLines(tabber, stats.UncoveredLines)
		}
	}

	fmt.Fprintf(tabber, "\n")
}

func reportMissingExplanations(w io.Writer, result AnalyzeResult) {
	if len(result.FilesWithMissingExplanations) == 0 {
		return
	}

	tabber := tabwriter.NewWriter(w, 1, 8, 2, '\t', 0) //nolint:mnd // relax
	defer tabber.Flush()

	fmt.Fprintf(tabber, "\nFiles with missing explanations for coverage-ignore annotations:")
	fmt.Fprintf(tabber, "\n  file:\tline numbers:")

	for _, stats := range result.FilesWithMissingExplanations {
		if len(stats.AnnotationsWithoutComments) == 0 {
			continue
		}

		fmt.Fprintf(tabber, "\n  %s\t", stats.Name)

		separator := ""
		for _, ann := range stats.AnnotationsWithoutComments {
			fmt.Fprintf(tabber, "%s%d", separator, ann)
			separator = ", "
		}
	}

	fmt.Fprintf(tabber, "\n")
}

//nolint:lll // relax
func reportDiff(w io.Writer, result AnalyzeResult) {
	if !result.HasBaseBreakdown {
		return
	}

	tabber := tabwriter.NewWriter(w, 1, 8, 2, '\t', 0) //nolint:mnd // relax
	defer tabber.Flush()

	if result.DiffThreshold != nil {
		status := statusStr(result.MeetsDiffThreshold())
		fmt.Fprintf(tabber, "\nCoverage difference threshold (%.2f%%) satisfied:\t %s", *result.DiffThreshold, status)
		fmt.Fprintf(tabber, "\nCoverage difference: %.2f%%\n", result.DiffPercentage)
	}

	if len(result.Diff) == 0 {
		fmt.Fprintf(tabber, "\nNo coverage changes in any files compared to the base.\n")
		return
	}

	td := TotalLinesMissingCoverage(result.Diff)
	fmt.Fprintf(tabber, "\nTest coverage has changed in the current files, with %d lines missing coverage.", td)
	fmt.Fprintf(tabber, "\n  file:\tuncovered:\tcurrent coverage:\tbase coverage:")

	for _, d := range result.Diff {
		var baseStr string
		if d.Base == nil {
			baseStr = " / "
		} else {
			baseStr = d.Base.Str()
		}

		dp := d.Current.UncoveredLinesCount()
		fmt.Fprintf(tabber, "\n  %s\t%3d\t%s\t%s", d.Current.Name, dp, d.Current.Str(), baseStr)
	}

	fmt.Fprintf(tabber, "\n")
}

func ReportForGithubAction(w io.Writer, result AnalyzeResult) { //nolint:maintidx // relax
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
			"%s: coverage: %s; threshold: %d%%",
			title, stats.Str(), stats.Threshold,
		)
		reportLineError(stats.Name, title, msg)
	}

	for _, stats := range result.PackagesBelowThreshold {
		title := "Package test coverage below threshold"
		msg := fmt.Sprintf(
			"%s: package: %s; coverage: %s; threshold: %d%%",
			title, stats.Name, stats.Str(), stats.Threshold,
		)
		reportError(title, msg)
	}

	if !result.MeetsTotalCoverage() {
		title := "Total test coverage below threshold"
		msg := fmt.Sprintf(
			"%s: coverage: %s; threshold: %d%%",
			title, result.TotalStats.Str(), result.Threshold.Total,
		)
		reportError(title, msg)
	}

	// Report missing explanations for coverage-ignore annotations
	for _, stats := range result.FilesWithMissingExplanations {
		if len(stats.AnnotationsWithoutComments) > 0 {
			for _, ann := range stats.AnnotationsWithoutComments {
				title := "Missing explanation for coverage-ignore"
				msg := title + ": add an explanation after the coverage-ignore annotation"

				file := stats.Name
				lineNumber := ann
				fmt.Fprintf(out, "::error file=%s,title=%s,line=%d::%s\n", file, title, lineNumber, msg)
			}
		}
	}
}

func reportGHWarning(out io.Writer, title, msg string) { // coverage-ignore
	fmt.Fprintf(out, "::warning title=%s::%s\n", title, msg)
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

	totalStr := strconv.Itoa(result.TotalStats.CoveredPercentage())

	return errors.Join(
		setOutputValue(file, gaOutputTotalCoverage, totalStr),
		setOutputValue(file, gaOutputBadgeColor, badge.Color(result.TotalStats.CoveredPercentage())),
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
	resp, _ := json.Marshal(s) //nolint:errcheck,errchkjson // relax
	return string(resp)
}

func compressUncoveredLines(w io.Writer, ull []int) {
	separator := ""
	printRange := func(a, b int) {
		if a == b {
			fmt.Fprintf(w, "%v%v", separator, a)
		} else {
			fmt.Fprintf(w, "%v%v-%v", separator, a, b)
		}

		separator = " "
	}

	last := -1
	for i := range ull {
		if last == -1 {
			last = ull[i]
		} else if ull[i-1]+1 != ull[i] {
			printRange(last, ull[i-1])
			last = ull[i]
		}
	}

	if last != -1 {
		printRange(last, ull[len(ull)-1])
	}
}

func statusStr(passing bool) string {
	if passing {
		return "PASS"
	}

	return "FAIL"
}
