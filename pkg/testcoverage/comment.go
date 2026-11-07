package testcoverage

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/google/go-github/v88/github"

	"github.com/vladopajic/go-test-coverage/v2/pkg/testcoverage/logger"
)

const (
	githubCommentPageSize  = 100
	githubCommentTimeout   = 30 * time.Second
	githubCommentBodyLimit = 65536
)
const coverageCommentMarker = `<!-- go-test-coverage "report" -->`

func postCoverageComment(w io.Writer, token, report string) error {
	options := []github.ClientOptionsFunc{github.WithAuthToken(token)}

	if apiURL := os.Getenv("GITHUB_API_URL"); apiURL != "" {
		apiURL = strings.TrimRight(apiURL, "/") + "/"
		options = append(options, github.WithURLs(&apiURL, nil))
	}

	client, err := github.NewClient(options...)
	if err != nil {
		return fmt.Errorf("create GitHub client: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), githubCommentTimeout)
	defer cancel()

	return postCoverageCommentWithClient(ctx, w, client, report)
}

func postCoverageCommentWithClient(
	ctx context.Context, w io.Writer, client *github.Client, report string,
) error {
	owner, repo, ok := strings.Cut(os.Getenv("GITHUB_REPOSITORY"), "/")
	if !ok || owner == "" || repo == "" {
		return errors.New("GITHUB_REPOSITORY must have format owner/repository")
	}

	number, err := coveragePullRequest(ctx, client, owner, repo)
	if err != nil {
		fmt.Fprintf(w, "Warning: pull request lookup failed; skipping coverage comment: %v\n", err)
		return nil
	}

	if number == 0 {
		logger.L.Info().Msg("No open pull request found; skipping coverage comment.")
		return nil
	}

	id, err := findCoverageComment(ctx, client, owner, repo, number)
	if err != nil {
		return err
	}

	body := coverageCommentBody(report)

	comment := &github.IssueComment{Body: &body}
	if id != 0 {
		_, _, err = client.Issues.EditComment(ctx, owner, repo, id, comment)
	} else {
		_, _, err = client.Issues.CreateComment(ctx, owner, repo, number, comment)
	}

	if err != nil {
		return fmt.Errorf("save coverage comment: %w", err)
	}

	return nil
}

func coverageCommentBody(report string) string {
	const (
		prefix          = "go-test-coverage report:\n```\n"
		suffix          = "```\n" + coverageCommentMarker
		truncatedSuffix = "\n```\nReport truncated; see the full report in the workflow logs.\n" +
			coverageCommentMarker
	)

	body := prefix + report + suffix
	if utf8.RuneCountInString(body) <= githubCommentBodyLimit {
		return body
	}

	limit := githubCommentBodyLimit - utf8.RuneCountInString(prefix+truncatedSuffix)

	return prefix + string([]rune(report)[:limit]) + truncatedSuffix
}

func findCoverageComment(
	ctx context.Context, client *github.Client, owner, repo string, number int,
) (int64, error) {
	authorID, err := coverageCommentAuthor(ctx, client)
	if err != nil {
		return 0, err
	}

	opts := &github.IssueListCommentsOptions{
		ListOptions: github.ListOptions{PerPage: githubCommentPageSize},
	}
	for {
		comments, response, err := client.Issues.ListComments(ctx, owner, repo, number, opts)
		if err != nil {
			return 0, fmt.Errorf("list coverage comments: %w", err)
		}

		for _, existing := range comments {
			if existing.GetUser().GetNodeID() == authorID &&
				strings.Contains(existing.GetBody(), coverageCommentMarker) {
				return existing.GetID(), nil
			}
		}

		if response.NextPage == 0 {
			return 0, nil
		}

		opts.Page = response.NextPage
	}
}

// GraphQL viewer supports installation tokens as well as personal access tokens.
func coverageCommentAuthor(ctx context.Context, client *github.Client) (string, error) {
	endpoint := os.Getenv("GITHUB_GRAPHQL_URL")
	if endpoint == "" {
		endpoint = "graphql"
	}

	req, err := client.NewRequest(ctx, "POST", endpoint,
		map[string]string{"query": "query { viewer { id } }"})
	if err != nil {
		return "", fmt.Errorf("create posting identity request: %w", err)
	}

	var result struct {
		Data struct {
			Viewer struct {
				ID string `json:"id"`
			} `json:"viewer"`
		} `json:"data"`
		Errors []json.RawMessage `json:"errors"`
	}
	if _, err = client.Do(req, &result); err != nil {
		return "", fmt.Errorf("get posting identity: %w", err)
	}

	if len(result.Errors) != 0 || result.Data.Viewer.ID == "" {
		return "", errors.New("could not verify coverage comment posting identity")
	}

	return result.Data.Viewer.ID, nil
}

func coveragePullRequest(
	ctx context.Context, client *github.Client, owner, repo string,
) (int, error) {
	switch os.Getenv("GITHUB_EVENT_NAME") {
	case "pull_request", "pull_request_target":
		//nolint:gosec // GitHub supplies the trusted runner event file path.
		data, err := os.ReadFile(os.Getenv("GITHUB_EVENT_PATH"))
		if err != nil {
			return 0, fmt.Errorf("read GitHub event: %w", err)
		}

		var event github.PullRequestEvent
		if err = json.Unmarshal(data, &event); err != nil {
			return 0, fmt.Errorf("decode GitHub event: %w", err)
		}

		return event.GetPullRequest().GetNumber(), nil
	case "push":
		if os.Getenv("GITHUB_REF_TYPE") != "branch" {
			return 0, nil
		}

		prs, _, err := client.PullRequests.List(ctx, owner, repo, &github.PullRequestListOptions{
			State: "open",
			Head:  owner + ":" + os.Getenv("GITHUB_REF_NAME"),
		})
		if err != nil {
			return 0, fmt.Errorf("find pull request: %w", err)
		}

		if len(prs) > 0 {
			return prs[0].GetNumber(), nil
		}
	}

	return 0, nil
}
