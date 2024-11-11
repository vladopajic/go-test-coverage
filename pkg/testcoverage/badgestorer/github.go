package badgestorer

import (
	"context"
	"fmt"
	"net/http"

	"github.com/google/go-github/v56/github"
)

type Git struct {
	Token      string
	Owner      string
	Repository string
	Branch     string
	FileName   string
}

func GitPublicURL(cfg Git) string {
	return fmt.Sprintf("https://raw.githubusercontent.com/%s/%s/%s/%s",
		cfg.Owner, cfg.Repository, cfg.Branch, cfg.FileName,
	)
}

type githubStorer struct {
	cfg Git
}

func NewGithub(cfg Git) Storer {
	return &githubStorer{cfg: cfg}
}

func (s *githubStorer) Store(data []byte) (bool, error) {
	git := s.cfg
	client := github.NewClient(nil).WithAuthToken(git.Token)

	updateBadge := func(sha *string) (bool, error) {
		_, _, err := client.Repositories.UpdateFile(
			context.Background(),
			git.Owner,
			git.Repository,
			git.FileName,
			&github.RepositoryContentFileOptions{
				Message: github.String("update badge " + git.FileName),
				Content: data,
				Branch:  &git.Branch,
				SHA:     sha,
			},
		)
		if err != nil {
			return false, fmt.Errorf("update badge contents: %w", err)
		}

		return true, nil // has changed
	}

	fc, _, httpResp, err := client.Repositories.GetContents(
		context.Background(),
		git.Owner,
		git.Repository,
		git.FileName,
		&github.RepositoryContentGetOptions{Ref: git.Branch},
	)
	if httpResp.StatusCode == http.StatusNotFound { // when badge is not found create it
		return updateBadge(nil)
	}

	if err != nil { // coverage-ignore
		return false, fmt.Errorf("get badge content: %w", err)
	}

	content, err := fc.GetContent()
	if err != nil { // coverage-ignore
		return false, fmt.Errorf("decode badge content: %w", err)
	}

	if content == string(data) { // same badge already exists... do nothing
		return false, nil
	}

	return updateBadge(fc.SHA)
}
