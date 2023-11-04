package testcoverage

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"os"

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

	buffer := &bytes.Buffer{}
	out := bufio.NewWriter(buffer)

	defer func() {
		out.Flush()

		if buffer.Len() != 0 {
			fmt.Fprintf(w, "\n-------------------------\n")
			w.Write(buffer.Bytes()) //nolint:errcheck // relx
		}
	}()

	if cfg.Badge.FileName != "" {
		err := saveBadeToFile(out, cfg.Badge.FileName, badge)
		if err != nil {
			return fmt.Errorf("save badge to file: %w", err)
		}
	}

	if cfg.Badge.CDN.Secret != "" {
		err := saveBadgeToCDN(out, cfg.Badge.CDN, badge)
		if err != nil {
			return fmt.Errorf("save badge to cdn: %w", err)
		}
	}

	if cfg.Badge.Git.Token != "" {
		err := saveBadgeToGitRepo(out, cfg.Badge.Git, badge)
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

	fmt.Fprintf(w, "Badge saved to file '%v'\n", filename)

	return nil
}

type Git struct {
	Token      string
	Owner      string
	Repository string
	Branch     string
	FileName   string
}

func saveBadgeToGitRepo(w io.Writer, git Git, data []byte) error {
	changed, err := updateGithubBadge(git, data)
	if err != nil {
		return err
	}

	if changed {
		fmt.Fprintf(w, "Badge with updated coverage pushed\n")
	} else {
		fmt.Fprintf(w, "Badge with same coverage already pushed (nothing to commit)\n")
	}

	fmt.Fprintf(w, "\nEmbed this badge with markdown:\n")
	fmt.Fprintf(w,
		"![coverage](https://raw.githubusercontent.com/%s/%s/%s/%s)\n",
		git.Owner, git.Repository, git.Branch, git.FileName,
	)

	return nil
}

func updateGithubBadge(git Git, data []byte) (bool, error) {
	client := github.NewClient(nil).WithAuthToken(git.Token)

	updateBadge := func(sha *string) (bool, error) {
		_, _, err := client.Repositories.UpdateFile(
			context.Background(),
			git.Owner,
			git.Repository,
			git.FileName,
			&github.RepositoryContentFileOptions{
				Message: github.String(fmt.Sprintf("update badge %s", git.FileName)),
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

	if err != nil {
		return false, fmt.Errorf("get badge content: %w", err)
	}

	content, err := fc.GetContent()
	if err != nil {
		return false, fmt.Errorf("decode badge content: %w", err)
	}

	if content == string(data) { // same badge already exists... do nothing
		return false, nil
	}

	return updateBadge(fc.SHA)
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
	changed, err := updateBadgeCDN(cdn, data)
	if err != nil {
		return err
	}

	if changed {
		fmt.Fprintf(w, "Badge with updated coverage uploaded to CDN. Badge path: %v\n", cdn.FileName)
	} else {
		fmt.Fprintf(w, "Badge with same coverage already uploaded to CDN.\n")
	}

	return nil
}

func updateBadgeCDN(cdn CDN, data []byte) (bool, error) {
	s3Client, err := createS3Client(cdn)
	if err != nil {
		return false, fmt.Errorf("create s3 client: %w", err)
	}

	// First get object and check if data differs that currently uploaded
	result, err := s3Client.GetObject(&s3.GetObjectInput{
		Bucket: aws.String(cdn.BucketName),
		Key:    aws.String(cdn.FileName),
	})
	if err == nil {
		resp, _ := io.ReadAll(result.Body)
		if bytes.Equal(resp, data) {
			return false, nil // has not changed
		}
	}

	// Currently uploaded badge does not exists or has changed
	// so it should be uploaded
	_, err = s3Client.PutObject(&s3.PutObjectInput{
		Bucket:        aws.String(cdn.BucketName),
		Key:           aws.String(cdn.FileName),
		Body:          bytes.NewReader(data),
		ContentType:   aws.String(badge.ContentType),
		ContentLength: aws.Int64(int64(len(data))),
	})
	if err != nil {
		return false, fmt.Errorf("put object: %w", err)
	}

	return true, nil // has changed
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
