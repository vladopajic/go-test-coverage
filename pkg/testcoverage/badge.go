package testcoverage

import (
	"bufio"
	"bytes"
	"fmt"
	"io"

	"github.com/subhambhardwaj/go-test-coverage/v2/pkg/testcoverage/badge"
	"github.com/subhambhardwaj/go-test-coverage/v2/pkg/testcoverage/badgestorer"
)

func generateAndSaveBadge(w io.Writer, cfg Config, totalCoverage int) error {
	badge, err := badge.Generate(totalCoverage)
	if err != nil { // coverage-ignore // should never happen
		return fmt.Errorf("generate badge: %w", err)
	}

	buffer := &bytes.Buffer{}
	out := bufio.NewWriter(buffer)

	// `out` writer is used as temporall buffer, which will be finally
	// written to `w` in this defer call
	defer func() {
		out.Flush()

		if buffer.Len() != 0 {
			// add visual separator before writing result
			// of generate and save badge action
			fmt.Fprintf(w, "\n-------------------------\n")
			w.Write(buffer.Bytes()) //nolint:errcheck // relax
		}
	}()

	return storeBadge(out, defaultStorerFactories(), cfg, badge)
}

type storerFactories struct {
	File func(string) badgestorer.Storer
	Git  func(badgestorer.Git) badgestorer.Storer
	CDN  func(badgestorer.CDN) badgestorer.Storer
}

func defaultStorerFactories() storerFactories {
	return storerFactories{
		File: badgestorer.NewFile,
		Git:  badgestorer.NewGithub,
		CDN:  badgestorer.NewCDN,
	}
}

func storeBadge(w io.Writer, sf storerFactories, config Config, badge []byte) error {
	if fn := config.Badge.FileName; fn != "" {
		_, err := sf.File(fn).Store(badge)
		if err != nil {
			return fmt.Errorf("save badge to file: %w", err)
		}

		fmt.Fprintf(w, "Badge saved to file '%v'\n", fn)
	}

	if cfg := config.Badge.CDN; cfg.Secret != "" {
		changed, err := sf.CDN(cfg).Store(badge)
		if err != nil {
			return fmt.Errorf("save badge to cdn: %w", err)
		}

		if changed {
			fmt.Fprintf(w, "Badge with updated coverage uploaded to CDN. Badge path: %v\n", cfg.FileName)
		} else {
			fmt.Fprintf(w, "Badge with same coverage already uploaded to CDN.\n")
		}
	}

	if cfg := config.Badge.Git; cfg.Token != "" {
		changed, err := sf.Git(cfg).Store(badge)
		if err != nil {
			return fmt.Errorf("save badge to git branch: %w", err)
		}

		if changed {
			fmt.Fprintf(w, "Badge with updated coverage pushed\n")
		} else {
			fmt.Fprintf(w, "Badge with same coverage already pushed (nothing to commit)\n")
		}

		fmt.Fprintf(w, "\nEmbed this badge with markdown:\n")
		fmt.Fprintf(w, "![coverage](%s)\n", badgestorer.GitPublicURL(cfg))
	}

	return nil
}
