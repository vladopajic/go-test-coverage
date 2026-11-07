package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"gopkg.in/yaml.v3"
)

//nolint:gocognit // Table-driven shell matrix covers both actions and token sources.
func TestActionTokenEnvironment(t *testing.T) {
	t.Parallel()

	bash, err := exec.LookPath("bash")
	if err != nil {
		t.Skip("bash is required to test the action shell scripts")
	}

	entrypoint, err := os.ReadFile("docker-entrypoint.sh")
	require.NoError(t, err)

	var source struct {
		Runs struct {
			Steps []struct {
				ID  string `yaml:"id"`
				Run string `yaml:"run"`
			} `yaml:"steps"`
		} `yaml:"runs"`
	}

	sourceAction, err := os.ReadFile("action/source/action.yml")
	require.NoError(t, err)
	require.NoError(t, yaml.Unmarshal(sourceAction, &source))

	var sourceSetup string

	for _, step := range source.Runs.Steps {
		if step.ID == "coverage" {
			sourceSetup, _, _ = strings.Cut(step.Run, "go run ")
		}
	}

	require.NotEmpty(t, sourceSetup)

	for name, script := range map[string]string{
		"docker": strings.Replace(string(entrypoint),
			"args=(/go-test-coverage)", "args=(\"$GTC_TEST_BINARY\")", 1),
		"source": sourceSetup + "exec \"$GTC_TEST_BINARY\"\n",
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			for _, tc := range []struct {
				name     string
				input    string
				present  bool
				existing string
				want     string
			}{
				{name: "unset input", existing: "existing-token", want: "existing-token"},
				{name: "empty input", present: true, existing: "existing-token", want: "existing-token"},
				{
					name: "input overrides environment", present: true, input: "input-token",
					existing: "existing-token", want: "input-token",
				},
				{name: "no token"},
			} {
				t.Run(tc.name, func(t *testing.T) {
					t.Parallel()

					dir := t.TempDir()
					mock := filepath.Join(dir, "mock")
					//nolint:gosec // The mock must be executable to exercise the entrypoint exec.
					require.NoError(t, os.WriteFile(mock,
						[]byte("#!/bin/bash\nprintf '%s' \"${GITHUB_TOKEN:-}\"\n"), 0o700))

					command := exec.CommandContext(t.Context(), bash, "-e", "-c", script)

					for _, env := range os.Environ() {
						if !strings.HasPrefix(env, "GITHUB_TOKEN=") &&
							!strings.HasPrefix(env, "INPUT_GITHUB_TOKEN=") {
							command.Env = append(command.Env, env)
						}
					}

					command.Env = append(command.Env, "GTC_TEST_BINARY="+filepath.ToSlash(mock),
						"GITHUB_TOKEN="+tc.existing, "INPUT_COMMENT=true")
					if tc.present {
						command.Env = append(command.Env, "INPUT_GITHUB_TOKEN="+tc.input)
					}

					output, err := command.CombinedOutput()
					require.NoError(t, err, "%s", output)
					assert.Equal(t, tc.want, string(output))
				})
			}
		})
	}
}
