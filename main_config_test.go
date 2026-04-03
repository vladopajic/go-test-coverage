package main

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/vladopajic/go-test-coverage/v2/pkg/testcoverage"
	"github.com/vladopajic/go-test-coverage/v2/pkg/testcoverage/badgestorer"
)

func ptr[T any](v T) *T { return &v }

//nolint:lll // realx
func Test_args_overrideConfig(t *testing.T) {
	t.Parallel()

	t.Run("no args leaves config unchanged", func(t *testing.T) {
		t.Parallel()

		cfg := testcoverage.Config{Profile: "cover.out"}
		result, err := (&args{}).overrideConfig(cfg)
		assert.NoError(t, err)
		assert.Equal(t, cfg, result)
	})

	t.Run("Profile", func(t *testing.T) {
		t.Parallel()

		result, err := (&args{Profile: ptr("new.out")}).overrideConfig(testcoverage.Config{})
		assert.NoError(t, err)
		assert.Equal(t, "new.out", result.Profile)
	})

	t.Run("Debug", func(t *testing.T) {
		t.Parallel()

		result, err := (&args{Debug: true}).overrideConfig(testcoverage.Config{})
		assert.NoError(t, err)
		assert.True(t, result.Debug)
	})

	t.Run("GithubActionOutput", func(t *testing.T) {
		t.Parallel()

		result, err := (&args{GithubActionOutput: true}).overrideConfig(testcoverage.Config{})
		assert.NoError(t, err)
		assert.True(t, result.GithubActionOutput)
	})

	t.Run("SourceDir", func(t *testing.T) {
		t.Parallel()

		result, err := (&args{SourceDir: ptr("./src")}).overrideConfig(testcoverage.Config{})
		assert.NoError(t, err)
		assert.Equal(t, "./src", result.SourceDir)
	})

	t.Run("ThresholdFile", func(t *testing.T) {
		t.Parallel()

		result, err := (&args{ThresholdFile: ptr(80)}).overrideConfig(testcoverage.Config{})
		assert.NoError(t, err)
		assert.Equal(t, 80, result.Threshold.File)
	})

	t.Run("ThresholdPackage", func(t *testing.T) {
		t.Parallel()

		result, err := (&args{ThresholdPackage: ptr(70)}).overrideConfig(testcoverage.Config{})
		assert.NoError(t, err)
		assert.Equal(t, 70, result.Threshold.Package)
	})

	t.Run("ThresholdTotal", func(t *testing.T) {
		t.Parallel()

		result, err := (&args{ThresholdTotal: ptr(90)}).overrideConfig(testcoverage.Config{})
		assert.NoError(t, err)
		assert.Equal(t, 90, result.Threshold.Total)
	})

	t.Run("BreakdownFileName", func(t *testing.T) {
		t.Parallel()

		result, err := (&args{BreakdownFileName: ptr("breakdown.out")}).overrideConfig(testcoverage.Config{})
		assert.NoError(t, err)
		assert.Equal(t, "breakdown.out", result.BreakdownFileName)
	})

	t.Run("DiffBaseBreakdownFileName", func(t *testing.T) {
		t.Parallel()

		result, err := (&args{DiffBaseBreakdownFileName: ptr("base.out")}).overrideConfig(testcoverage.Config{})
		assert.NoError(t, err)
		assert.Equal(t, "base.out", result.Diff.BaseBreakdownFileName)
	})

	t.Run("BadgeFileName", func(t *testing.T) {
		t.Parallel()

		result, err := (&args{BadgeFileName: ptr("badge.svg")}).overrideConfig(testcoverage.Config{})
		assert.NoError(t, err)
		assert.Equal(t, "badge.svg", result.Badge.FileName)
	})

	t.Run("CDN secret with all fields", func(t *testing.T) {
		t.Parallel()

		a := &args{
			CDNSecret:         ptr("secret"),
			CDNKey:            ptr("key"),
			CDNRegion:         ptr("us-east-1"),
			CDNFileName:       ptr("badge.svg"),
			CDNBucketName:     ptr("my-bucket"),
			CDNEndpoint:       ptr("https://s3.example.com"),
			CDNForcePathStyle: true,
		}
		result, err := a.overrideConfig(testcoverage.Config{})
		assert.NoError(t, err)
		assert.Equal(t, badgestorer.CDN{
			Secret:         "secret",
			Key:            "key",
			Region:         "us-east-1",
			FileName:       "badge.svg",
			BucketName:     "my-bucket",
			Endpoint:       "https://s3.example.com",
			ForcePathStyle: true,
		}, result.Badge.CDN)
	})

	t.Run("CDN secret with nil optional fields", func(t *testing.T) {
		t.Parallel()

		a := &args{CDNSecret: ptr("secret")}
		result, err := a.overrideConfig(testcoverage.Config{})
		assert.NoError(t, err)
		assert.Equal(t, "secret", result.Badge.CDN.Secret)
		assert.Empty(t, result.Badge.CDN.Key)
		assert.Empty(t, result.Badge.CDN.Endpoint)
	})

	t.Run("CDN not set when secret is nil", func(t *testing.T) {
		t.Parallel()

		a := &args{CDNKey: ptr("key")}
		result, err := a.overrideConfig(testcoverage.Config{})
		assert.NoError(t, err)
		assert.Empty(t, result.Badge.CDN.Secret)
		assert.Empty(t, result.Badge.CDN.Key)
	})

	t.Run("Git token with valid repository", func(t *testing.T) {
		t.Parallel()

		a := &args{
			GitToken:      ptr("token"),
			GitRepository: ptr("owner/repo"),
			GitBranch:     ptr("main"),
			GitFileName:   ptr("badge.svg"),
		}
		result, err := a.overrideConfig(testcoverage.Config{})
		assert.NoError(t, err)
		assert.Equal(t, badgestorer.Git{
			Token:      "token",
			Owner:      "owner",
			Repository: "repo",
			Branch:     "main",
			FileName:   "badge.svg",
		}, result.Badge.Git)
	})

	t.Run("Git token with nil optional fields", func(t *testing.T) {
		t.Parallel()

		a := &args{
			GitToken:      ptr("token"),
			GitRepository: ptr("owner/repo"),
		}
		result, err := a.overrideConfig(testcoverage.Config{})
		assert.NoError(t, err)
		assert.Equal(t, "token", result.Badge.Git.Token)
		assert.Equal(t, "owner", result.Badge.Git.Owner)
		assert.Equal(t, "repo", result.Badge.Git.Repository)
		assert.Empty(t, result.Badge.Git.Branch)
		assert.Empty(t, result.Badge.Git.FileName)
	})

	t.Run("Git token with invalid repository format", func(t *testing.T) {
		t.Parallel()

		a := &args{
			GitToken:      ptr("token"),
			GitRepository: ptr("invalid-no-slash"),
		}
		_, err := a.overrideConfig(testcoverage.Config{})
		assert.Error(t, err)
	})

	t.Run("Git not set when token is nil", func(t *testing.T) {
		t.Parallel()

		a := &args{GitRepository: ptr("owner/repo")}
		result, err := a.overrideConfig(testcoverage.Config{})
		assert.NoError(t, err)
		assert.Empty(t, result.Badge.Git.Token)
		assert.Empty(t, result.Badge.Git.Owner)
	})

	t.Run("args do not override existing config values when nil", func(t *testing.T) {
		t.Parallel()

		cfg := testcoverage.Config{
			Profile:           "original.out",
			BreakdownFileName: "original-breakdown.out",
			Threshold: testcoverage.Threshold{
				File:    10,
				Package: 20,
				Total:   30,
			},
		}
		result, err := (&args{}).overrideConfig(cfg)
		assert.NoError(t, err)
		assert.Equal(t, cfg, result)
	})
}
