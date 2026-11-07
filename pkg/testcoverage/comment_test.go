package testcoverage_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	. "github.com/vladopajic/go-test-coverage/v2/pkg/testcoverage"
)

// These tests exercise the full check against a local GitHub API, including reports
// posted before a threshold failure. Environment variables prevent parallel execution.
//
//nolint:cyclop,gocognit,gocyclo // Table-driven mock API exercises multiple request sequences.
func TestCoverageComment(t *testing.T) {
	if testing.Short() {
		t.Skip("coverage comment tests run full coverage checks against a mock GitHub API")
	}

	const marker = `<!-- go-test-coverage "report" -->`

	for _, tc := range []struct {
		name          string
		event         string
		refType       string
		disabled      bool
		noPR          bool
		update        bool
		paginate      bool
		failStage     string
		threshold     int
		wantWrites    int
		foreign       bool
		missingAuthor bool
	}{
		{name: "create on pull request", event: "pull_request", wantWrites: 1},
		{name: "create on branch push", event: "push", refType: "branch", wantWrites: 1},
		{name: "update existing report", event: "pull_request", update: true, wantWrites: 1},
		{
			name: "find report on second page", event: "pull_request",
			update: true, paginate: true, wantWrites: 1,
		},
		{name: "ignore foreign report", event: "pull_request", foreign: true, wantWrites: 1},
		{name: "ignore report without author", event: "pull_request", missingAuthor: true, wantWrites: 1},
		{
			name: "skip foreign report before own report", event: "pull_request", foreign: true,
			update: true, wantWrites: 1,
		},
		{name: "identity request error", event: "pull_request", failStage: "identity-request"},
		{name: "identity HTTP error", event: "pull_request", failStage: "identity"},
		{name: "identity GraphQL error", event: "pull_request", failStage: "graphql"},
		{name: "identity missing", event: "pull_request", failStage: "missing-identity"},
		{name: "post failing coverage", event: "pull_request", threshold: 100, wantWrites: 1},
		{name: "disabled", event: "pull_request", disabled: true},
		{name: "no open pull request", event: "push", refType: "branch", noPR: true},
		{name: "tag push", event: "push", refType: "tag"},
		{name: "unsupported event", event: "workflow_dispatch"},
		{name: "lookup error", event: "push", refType: "branch", failStage: "lookup"},
		{name: "list error", event: "pull_request", failStage: "list"},
		{name: "create error", event: "pull_request", failStage: "write", wantWrites: 1},
		{name: "update error", event: "pull_request", update: true, failStage: "write", wantWrites: 1},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var calls, writes int

			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				calls++

				assert.Equal(t, "Bearer test-token", r.Header.Get("Authorization"))
				w.Header().Set("Content-Type", "application/json")

				switch {
				case r.URL.Path == "/graphql":
					assert.Equal(t, http.MethodPost, r.Method)

					switch tc.failStage {
					case "identity":
						http.Error(w, `{}`, http.StatusForbidden)
					case "graphql":
						fmt.Fprint(w, `{"data":{"viewer":{"id":"poster-id"}},"errors":[{"message":"denied"}]}`)
					case "missing-identity":
						fmt.Fprint(w, `{"data":{"viewer":null}}`)
					default:
						fmt.Fprint(w, `{"data":{"viewer":{"id":"poster-id"}}}`)
					}
				case r.URL.Path == "/repos/owner/repo/pulls":
					assert.Equal(t, "open", r.URL.Query().Get("state"))
					assert.Equal(t, "owner:feature/test", r.URL.Query().Get("head"))

					switch {
					case tc.failStage == "lookup":
						http.Error(w, `{}`, http.StatusForbidden)
					case tc.noPR:
						fmt.Fprint(w, `[]`)
					default:
						fmt.Fprint(w, `[{"number":42}]`)
					}
				case r.Method == http.MethodGet && r.URL.Path == "/repos/owner/repo/issues/42/comments":
					switch tc.failStage {
					case "list":
						http.Error(w, `{}`, http.StatusForbidden)
					default:
						comments := []map[string]any{{
							"id": 3, "body": "unrelated comment", "user": map[string]string{"node_id": "poster-id"},
						}}
						if tc.foreign || tc.paginate {
							comments = append(comments, map[string]any{
								"id": 1, "body": marker,
								"user": map[string]string{
									"node_id": "other-id", "login": "github-actions[bot]", "type": "Bot",
								},
							})
							comments = append(comments, map[string]any{
								"id": 4, "body": marker,
								"user": map[string]string{
									"node_id": "participant-id", "login": "participant", "type": "User",
								},
							})
						}

						if tc.missingAuthor {
							comments = append(comments, map[string]any{"id": 2, "body": marker})
						}

						if tc.paginate && r.URL.Query().Get("page") == "" {
							w.Header().Set("Link", fmt.Sprintf(`<http://%s%s?page=2>; rel="next"`, r.Host, r.URL.Path))
						} else if tc.update {
							comments = append(comments, map[string]any{
								"id": 99, "body": "old report\n" + marker,
								"user": map[string]string{"node_id": "poster-id"},
							})
						}

						data, err := json.Marshal(comments)
						assert.NoError(t, err)
						_, err = w.Write(data)
						assert.NoError(t, err)
					}

				case r.Method == http.MethodPost || r.Method == http.MethodPatch:
					writes++

					if tc.update {
						assert.Equal(t, http.MethodPatch, r.Method)
						assert.Equal(t, "/repos/owner/repo/issues/comments/99", r.URL.Path)
					} else {
						assert.Equal(t, http.MethodPost, r.Method)
						assert.Equal(t, "/repos/owner/repo/issues/42/comments", r.URL.Path)
					}

					var comment struct {
						Body string `json:"body"`
					}
					assert.NoError(t, json.NewDecoder(r.Body).Decode(&comment))
					assert.Contains(t, comment.Body, "go-test-coverage report:\n```\n")
					assert.Contains(t, comment.Body, "Total test coverage:")
					assert.Contains(t, comment.Body, marker)

					if tc.threshold == 100 {
						assert.Contains(t, comment.Body, "FAIL")
					}

					if tc.failStage == "write" {
						http.Error(w, `{}`, http.StatusForbidden)
					} else {
						fmt.Fprint(w, `{"id":99}`)
					}
				default:
					t.Errorf("unexpected GitHub request: %s %s", r.Method, r.URL)
					http.NotFound(w, r)
				}
			}))
			t.Cleanup(server.Close)

			eventFile := filepath.Join(t.TempDir(), "event.json")
			require.NoError(t, os.WriteFile(eventFile, []byte(`{"pull_request":{"number":42}}`), 0o600))
			t.Setenv("GITHUB_API_URL", server.URL+"/")

			graphqlURL := ""
			if tc.event == "push" {
				graphqlURL = server.URL + "/graphql"
			}

			if tc.failStage == "identity-request" {
				graphqlURL = ":"
			}

			t.Setenv("GITHUB_GRAPHQL_URL", graphqlURL)
			t.Setenv("GITHUB_REPOSITORY", "owner/repo")
			t.Setenv("GITHUB_EVENT_NAME", tc.event)
			t.Setenv("GITHUB_EVENT_PATH", eventFile)
			t.Setenv("GITHUB_REF_TYPE", tc.refType)
			t.Setenv("GITHUB_REF_NAME", "feature/test")

			var out bytes.Buffer

			pass, err := Check(&out, Config{
				Profile: profileOK, SourceDir: sourceDir, Format: ShowDefault,
				GithubComment: !tc.disabled, GithubToken: "test-token",
				Threshold: Threshold{Total: tc.threshold},
			})
			if tc.failStage != "" && tc.failStage != "lookup" {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}

			assert.Equal(t, tc.threshold == 0 && (tc.failStage == "" || tc.failStage == "lookup"), pass)
			assert.Equal(t, tc.wantWrites, writes)

			if tc.disabled || tc.refType == "tag" || tc.event == "workflow_dispatch" {
				assert.Zero(t, calls)
			}
		})
	}
}

func TestCoverageCommentInvalidEnvironment(t *testing.T) {
	if testing.Short() {
		t.Skip("coverage comment tests run full coverage checks against a mock GitHub API")
	}

	for _, tc := range []struct {
		name         string
		repository   string
		apiURL       string
		event        string
		missingEvent bool
		wantError    string
		wantWarning  string
	}{
		{
			name: "invalid API URL", repository: "owner/repo", apiURL: ":",
			wantError: "create GitHub client",
		},
		{name: "missing repository", wantError: "GITHUB_REPOSITORY must have format"},
		{
			name: "repository without separator", repository: "repo",
			wantError: "GITHUB_REPOSITORY must have format",
		},
		{
			name: "missing repository owner", repository: "/repo",
			wantError: "GITHUB_REPOSITORY must have format",
		},
		{
			name: "missing repository name", repository: "owner/",
			wantError: "GITHUB_REPOSITORY must have format",
		},
		{
			name: "missing event file", repository: "owner/repo", missingEvent: true,
			wantWarning: "read GitHub event",
		},
		{
			name: "malformed event JSON", repository: "owner/repo", event: "{",
			wantWarning: "decode GitHub event",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				t.Errorf("unexpected GitHub request: %s %s", r.Method, r.URL)
				http.NotFound(w, r)
			}))
			t.Cleanup(server.Close)

			apiURL := server.URL + "/"
			if tc.apiURL != "" {
				apiURL = tc.apiURL
			}

			eventFile := filepath.Join(t.TempDir(), "event.json")
			if !tc.missingEvent {
				require.NoError(t, os.WriteFile(eventFile, []byte(tc.event), 0o600))
			}

			t.Setenv("GITHUB_API_URL", apiURL)
			t.Setenv("GITHUB_REPOSITORY", tc.repository)
			t.Setenv("GITHUB_EVENT_NAME", "pull_request")
			t.Setenv("GITHUB_EVENT_PATH", eventFile)

			var out bytes.Buffer

			pass, err := Check(&out, Config{
				Profile: profileOK, SourceDir: sourceDir, Format: ShowDefault,
				GithubComment: true, GithubToken: "test-token",
			})
			if tc.wantError != "" {
				require.ErrorContains(t, err, tc.wantError)
				assert.False(t, pass)
			} else {
				require.NoError(t, err)
				assert.True(t, pass)
				assert.Contains(t, out.String(), "Warning: pull request lookup failed")
				assert.Contains(t, out.String(), tc.wantWarning)
			}
		})
	}
}

func TestCommentConfig(t *testing.T) {
	t.Parallel()

	cfg := Config{Profile: "cover.out", GithubComment: true}
	assert.ErrorContains(t, cfg.Validate(), "github-token is required")
	cfg.GithubToken = "secret"
	assert.NoError(t, cfg.Validate())
	assert.Equal(t, HiddenValue, cfg.Redacted().GithubToken)
	assert.Equal(t, "secret", cfg.GithubToken)
}
