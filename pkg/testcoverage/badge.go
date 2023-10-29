package testcoverage

import (
	"fmt"
	"os"

	"github.com/vladopajic/go-test-coverage/v2/pkg/testcoverage/badge"
)

func GenerateAndSaveBadge(cfg Config, totalCoverage int) error {
	badge, err := badge.Generate(totalCoverage)
	if err != nil {
		return fmt.Errorf("generate badge: %w", err)
	}

	if cfg.Badge.FileName != "" {
		err := saveBadeToFile(cfg.Badge.FileName, badge)
		if err != nil {
			return fmt.Errorf("save badge to file: %w", err)
		}
	}

	return nil
}

func saveBadeToFile(filename string, data []byte) error {
	return os.WriteFile(filename, data, 0o644) //nolint:gosec,gomnd,wrapcheck // relax
}
