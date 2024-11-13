package badgestorer_test

import (
	"context"
	"os"
	"testing"

	"github.com/google/go-github/v56/github"
	"github.com/stretchr/testify/assert"
	"golang.org/x/exp/rand"

	. "github.com/vladopajic/go-test-coverage/v2/pkg/testcoverage/badgestorer"
)

const envGitToken = "GITHUB_TOKEN" //nolint:gosec // false-positive

func Test_Github_Error(t *testing.T) {
	t.Parallel()

	if testing.Short() {
		return
	}

	data := []byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 0}
	cfg := Git{
		Token:      `🔑`,
		Owner:      "owner",
		Repository: "repo",
	}
	s := NewGithub(cfg)

	updated, err := s.Store(data)
	assert.Error(t, err)
	assert.False(t, updated)
}

func Test_Github(t *testing.T) {
	t.Parallel()

	if testing.Short() {
		return
	}

	if getEnv(envGitToken) == "" {
		t.Skipf("%v env variable not set", envGitToken)
		return
	}

	data := []byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 0}
	cfg := Git{
		Token:      getEnv(envGitToken),
		Owner:      "vladopajic",
		Repository: "go-test-coverage",
		Branch:     "badges-integration-test",
		// random badge name must be used because two tests running from different platforms
		// in CI can cause race condition if badge has the same name
		FileName: "badge_" + randString() + ".svg",
	}
	s := NewGithub(cfg)

	// put badge
	updated, err := s.Store(data)
	assert.NoError(t, err)
	assert.True(t, updated)

	// put badge again - no change
	updated, err = s.Store(data)
	assert.NoError(t, err)
	assert.False(t, updated)

	// put badge again - expect change
	updated, err = s.Store(append(data, byte(1)))
	assert.NoError(t, err)
	assert.True(t, updated)

	deleteFile(t, cfg)
}

func getEnv(key string) string {
	value, _ := os.LookupEnv(key)
	return value
}

func deleteFile(t *testing.T, cfg Git) {
	t.Helper()

	client := github.NewClient(nil).WithAuthToken(cfg.Token)

	fc, _, _, err := client.Repositories.GetContents(
		context.Background(),
		cfg.Owner,
		cfg.Repository,
		cfg.FileName,
		&github.RepositoryContentGetOptions{Ref: cfg.Branch},
	)
	assert.NoError(t, err)

	_, _, err = client.Repositories.DeleteFile(
		context.Background(),
		cfg.Owner,
		cfg.Repository,
		cfg.FileName,
		&github.RepositoryContentFileOptions{
			Message: github.String("delete testing badge " + cfg.FileName),
			Branch:  &cfg.Branch,
			SHA:     fc.SHA,
		},
	)
	assert.NoError(t, err)
}

func randString() string {
	letterRunes := []rune("abcdefghijklmnopqrstuvwxyz")
	l := len(letterRunes)

	b := make([]rune, rand.Intn(10))
	for i := range b {
		b[i] = letterRunes[rand.Intn(l)]
	}

	return string(b)
}