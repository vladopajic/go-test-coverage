package testcoverage_test

import (
	"bytes"
	"io"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"

	. "github.com/subhambhardwaj/go-test-coverage/v2/pkg/testcoverage"
	"github.com/subhambhardwaj/go-test-coverage/v2/pkg/testcoverage/badge"
	"github.com/subhambhardwaj/go-test-coverage/v2/pkg/testcoverage/badgestorer"
)

func Test_GenerateAndSaveBadge_NoAction(t *testing.T) {
	t.Parallel()

	// Empty config - no action
	err := GenerateAndSaveBadge(nil, Config{}, 100)
	assert.NoError(t, err)
}

func Test_GenerateAndSaveBadge_SaveToFile(t *testing.T) {
	t.Parallel()

	if testing.Short() {
		return
	}

	const coverage = 100

	testFile := t.TempDir() + "/badge.svg"
	buf := &bytes.Buffer{}
	err := GenerateAndSaveBadge(buf, Config{
		Badge: Badge{
			FileName: testFile,
		},
	}, coverage)
	assert.NoError(t, err)
	assert.Contains(t, buf.String(), "Badge saved to file")

	contentBytes, err := os.ReadFile(testFile)
	assert.NoError(t, err)
	assert.NotEmpty(t, contentBytes)

	badge, err := badge.Generate(coverage)
	assert.NoError(t, err)
	assert.Equal(t, badge, contentBytes)
}

func Test_StoreBadge(t *testing.T) {
	t.Parallel()

	badge, err := badge.Generate(100)
	assert.NoError(t, err)

	someError := io.ErrShortBuffer

	// badge saved to file
	buf := &bytes.Buffer{}
	config := Config{Badge: Badge{
		FileName: t.TempDir() + "/badge.svg",
	}}
	sf := StorerFactories{File: fileFact(newStorer(true, nil))}
	err = StoreBadge(buf, sf, config, badge)
	assert.NoError(t, err)
	assert.Contains(t, buf.String(), "Badge saved to file")

	// failed to save badge
	buf = &bytes.Buffer{}
	sf = StorerFactories{File: fileFact(newStorer(false, someError))}
	err = StoreBadge(buf, sf, config, badge)
	assert.Error(t, err)
	assert.Empty(t, buf.String())

	// badge saved to cdn
	buf = &bytes.Buffer{}
	config = Config{Badge: Badge{
		CDN: badgestorer.CDN{Secret: `🔑`},
	}}
	sf = StorerFactories{CDN: cdnFact(newStorer(true, nil))}
	err = StoreBadge(buf, sf, config, badge)
	assert.NoError(t, err)
	assert.Contains(t, buf.String(), "Badge with updated coverage uploaded to CDN")

	// badge saved to cdn (no change)
	buf = &bytes.Buffer{}
	sf = StorerFactories{CDN: cdnFact(newStorer(false, nil))}
	err = StoreBadge(buf, sf, config, badge)
	assert.NoError(t, err)
	assert.Contains(t, buf.String(), "Badge with same coverage already uploaded to CDN")

	// failed to save cdn
	buf = &bytes.Buffer{}
	sf = StorerFactories{CDN: cdnFact(newStorer(false, someError))}
	err = StoreBadge(buf, sf, config, badge)
	assert.Error(t, err)
	assert.Empty(t, buf.String())

	// badge saved to git
	buf = &bytes.Buffer{}
	config = Config{Badge: Badge{
		Git: badgestorer.Git{Token: `🔑`},
	}}
	sf = StorerFactories{Git: gitFact(newStorer(true, nil))}
	err = StoreBadge(buf, sf, config, badge)
	assert.NoError(t, err)
	assert.Contains(t, buf.String(), "Badge with updated coverage pushed")

	// badge saved to git (no change)
	buf = &bytes.Buffer{}
	sf = StorerFactories{Git: gitFact(newStorer(false, nil))}
	err = StoreBadge(buf, sf, config, badge)
	assert.NoError(t, err)
	assert.Contains(t, buf.String(), "Badge with same coverage already pushed")

	// failed to save git
	buf = &bytes.Buffer{}
	sf = StorerFactories{Git: gitFact(newStorer(false, someError))}
	err = StoreBadge(buf, sf, config, badge)
	assert.Error(t, err)
	assert.Empty(t, buf.String())

	// save badge to all methods
	buf = &bytes.Buffer{}
	config = Config{Badge: Badge{
		FileName: t.TempDir() + "/badge.svg",
		Git:      badgestorer.Git{Token: `🔑`},
		CDN:      badgestorer.CDN{Secret: `🔑`},
	}}
	sf = StorerFactories{
		File: fileFact(newStorer(true, nil)),
		Git:  gitFact(newStorer(true, nil)),
		CDN:  cdnFact(newStorer(true, nil)),
	}
	err = StoreBadge(buf, sf, config, badge)
	assert.NoError(t, err)
	assert.Contains(t, buf.String(), "Badge saved to file")
	assert.Contains(t, buf.String(), "Badge with updated coverage pushed")
	assert.Contains(t, buf.String(), "Badge with updated coverage uploaded to CDN")
}

func fileFact(s badgestorer.Storer) func(string) badgestorer.Storer {
	return func(_ string) badgestorer.Storer {
		return s
	}
}

func cdnFact(s badgestorer.Storer) func(badgestorer.CDN) badgestorer.Storer {
	return func(_ badgestorer.CDN) badgestorer.Storer {
		return s
	}
}

func gitFact(s badgestorer.Storer) func(badgestorer.Git) badgestorer.Storer {
	return func(_ badgestorer.Git) badgestorer.Storer {
		return s
	}
}

func newStorer(updated bool, err error) badgestorer.Storer {
	return mockStorer{updated, err}
}

type mockStorer struct {
	updated bool
	err     error
}

func (s mockStorer) Store([]byte) (bool, error) {
	return s.updated, s.err
}
