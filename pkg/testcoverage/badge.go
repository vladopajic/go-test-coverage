package testcoverage

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/google/go-github/v56/github"

	"github.com/vladopajic/go-test-coverage/v2/pkg/testcoverage/badge"
)

func GenerateAndSaveBadge(w io.Writer, cfg Config, totalCoverage int) error {
	badge, err := badge.Generate(totalCoverage)
	if err != nil {
		return fmt.Errorf("generate badge: %w", err)
	}

	if cfg.Badge.FileName != "" {
		err := saveBadeToFile(w, cfg.Badge.FileName, badge)
		if err != nil {
			return fmt.Errorf("save badge to file: %w", err)
		}
	}

	if cfg.Badge.CDN.Secret != "" {
		err := saveBadgeToCDN(w, cfg.Badge.CDN, badge)
		if err != nil {
			return fmt.Errorf("save badge to cdn: %w", err)
		}
	}

	if cfg.Badge.Git.Token != "" {
		err := saveBadgeToBranch(w, cfg.Badge.Git, badge)
		if err != nil {
			return fmt.Errorf("save badge to git branch: %w", err)
		}
	}

	return nil
}

//nolint:gosec,gomnd,wrapcheck // relax
func saveBadeToFile(w io.Writer, filename string, data []byte) error {
	err := os.WriteFile(filename, data, 0o644)
	if err != nil {
		return err
	}

	fmt.Fprintf(w, "\nBadge saved to file '%v'\n", filename)

	return nil
}

type Git struct {
	Token      string
	Repository string
	Branch     string
	FileName   string
}

//nolint:lll // relax
func updateGithubBadge(git Git, owner, repo, path string, data []byte) (bool, error) {
	ctx := context.TODO()
	client := github.NewClient(nil).WithAuthToken(git.Token)

	updateBadge := func(fc *github.RepositoryContent) (bool, error) {
		var sha *string
		if fc != nil {
			sha = fc.SHA
		}

		_, _, err := client.Repositories.UpdateFile(ctx, owner, repo, path, &github.RepositoryContentFileOptions{
			Message: github.String("badge update"),
			Content: data,
			Branch:  &git.Branch,
			SHA:     sha,
		})
		if err != nil {
			return false, fmt.Errorf("update badge contents: %w", err)
		}

		return true, nil
	}

	fc, _, httpResp, err := client.Repositories.GetContents(ctx, owner, repo, path, &github.RepositoryContentGetOptions{
		Ref: git.Branch,
	})
	if httpResp.StatusCode == http.StatusNotFound { // when badge is not found create it
		return updateBadge(nil)
	}

	if err != nil {
		return false, fmt.Errorf("get badge contents: %w", err)
	}

	content, err := fc.GetContent()
	if err != nil {
		return false, fmt.Errorf("decode badge contents: %w", err)
	}

	if content == string(data) { // same badge already exists... do nothing
		return false, nil
	}

	return updateBadge(fc)
}

func saveBadgeToBranch(w io.Writer, git Git, data []byte) error {
	repoParts := strings.Split(git.Repository, "/")
	owner, repo := repoParts[0], repoParts[1]
	path := git.FileName

	changed, err := updateGithubBadge(git, owner, repo, path, data)
	if err != nil {
		return err
	}

	if changed {
		fmt.Fprintf(w, "\nBadge pushed to branch\n")
	} else {
		fmt.Fprintf(w, "\nBadge with same coverage already pushed - nothing to commit\n")
	}

	fmt.Fprintf(w, "\nEmbed this badge with markdown:\n")
	fmt.Fprintf(w,
		"![coverage](https://raw.githubusercontent.com/%s/%s/%s/%s)\n",
		owner, repo, git.Branch, git.FileName,
	)

	return nil
}

type CDN struct {
	Key            string
	Secret         string
	Region         string
	FileName       string
	BucketName     string
	Endpoint       string
	ForcePathStyle bool
}

func saveBadgeToCDN(w io.Writer, cdn CDN, data []byte) error {
	s3Client, err := createS3Client(cdn)
	if err != nil {
		return fmt.Errorf("create s3 client: %w", err)
	}

	object := s3.PutObjectInput{
		Bucket:        aws.String(cdn.BucketName),
		Key:           aws.String(cdn.FileName),
		Body:          bytes.NewReader(data),
		ContentType:   aws.String(badge.ContentType),
		ContentLength: aws.Int64(int64(len(data))),
	}

	_, err = s3Client.PutObject(&object)
	if err != nil {
		return fmt.Errorf("put object: %w", err)
	}

	fmt.Fprintf(w, "\nBadge uploaded to CDN\n")

	return nil
}

func createS3Client(cdn CDN) (*s3.S3, error) {
	s3Config := &aws.Config{
		Credentials:      credentials.NewStaticCredentials(cdn.Key, cdn.Secret, ""),
		Endpoint:         aws.String(cdn.Endpoint),
		Region:           aws.String(cdn.Region),
		S3ForcePathStyle: aws.Bool(cdn.ForcePathStyle),
	}

	newSession, err := session.NewSession(s3Config)
	if err != nil {
		return nil, fmt.Errorf("create session: %w", err)
	}

	s3Client := s3.New(newSession)

	return s3Client, nil
}
