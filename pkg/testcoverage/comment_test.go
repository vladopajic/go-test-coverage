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
		name       string
		event      string
		refType    string
		disabled   bool
		noPR       bool
		update     bool
		paginate   bool
		failStage  string
		threshold  int
		wantWrites int
	}{
		{name: "create on pull request", event: "pull_request", wantWrites: 1},
		{name: "create on branch push", event: "push", refType: "branch", wantWrites: 1},
		{name: "update existing report", event: "pull_request", update: true, wantWrites: 1},
		{
			name: "find report on second page", event: "pull_request",
			update: true, paginate: true, wantWrites: 1,
		},
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
					switch {
					case tc.failStage == "list":
						http.Error(w, `{}`, http.StatusForbidden)
					case tc.paginate && r.URL.Query().Get("page") == "":
						w.Header().Set("Link", fmt.Sprintf(`<http://%s%s?page=2>; rel="next"`, r.Host, r.URL.Path))
						fmt.Fprint(w, `[{"id":1,"body":"unrelated"}]`)
					case tc.update:
						data, err := json.Marshal([]map[string]any{{"id": 99, "body": "old report\n" + marker}})
						assert.NoError(t, err)
						_, err = w.Write(data)
						assert.NoError(t, err)
					default:
						fmt.Fprint(w, `[]`)
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
			t.Setenv("GITHUB_REPOSITORY", "owner/repo")
			t.Setenv("GITHUB_EVENT_NAME", tc.event)
			t.Setenv("GITHUB_EVENT_PATH", eventFile)
			t.Setenv("GITHUB_REF_TYPE", tc.refType)
			t.Setenv("GITHUB_REF_NAME", "feature/test")

			var out bytes.Buffer

			pass, err := Check(&out, Config{
				Profile: profileOK, SourceDir: sourceDir, Format: ShowDefault,
				Comment: !tc.disabled, GithubToken: "test-token",
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

func TestCommentConfig(t *testing.T) {
	t.Parallel()

	cfg := Config{Profile: "cover.out", Comment: true}
	assert.ErrorContains(t, cfg.Validate(), "github-token is required")
	cfg.GithubToken = "secret"
	assert.NoError(t, cfg.Validate())
	assert.Equal(t, HiddenValue, cfg.Redacted().GithubToken)
	assert.Equal(t, "secret", cfg.GithubToken)
}
