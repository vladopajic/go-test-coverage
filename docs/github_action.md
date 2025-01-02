# go-test-coverage GitHub Action

The `go-test-coverage` GitHub Action provides the following capabilities:
- Enforce success of GitHub workflows only when a specified coverage threshold is met.
- Generate a coverage badge to display the total test coverage.
- Post a detailed coverage report as a comment on pull requests, including:
  - current test coverage
  - the difference compared to the base branch

## Action Inputs and Outputs

Action inputs and outputs are documented in [action.yml](/action.yml) file.


## Basic Usage

Here’s an example of how to integrate `go-test-coverage` in a GitHub workflow that uses a config file. This is the preferred way because the same config file can be used for running coverage checks locally.

```yml
  - name: check test coverage
    uses: vladopajic/go-test-coverage@v2
    with:
      config: ./.testcoverage.yml
```

Alternatively, if you don't need advanced configuration options from a config file, you can specify thresholds directly in the action properties.

```yml
  - name: check test coverage
    uses: vladopajic/go-test-coverage@v2
    with:
      profile: cover.out
      local-prefix: github.com/org/project
      threshold-file: 80
      threshold-package: 80
      threshold-total: 95
```

Note: When using a config file alongside action properties, specifying these parameters will override the corresponding values in the config file.

## Liberal Coverage Check

The `go-test-coverage` GitHub Action can be configured to report the current test coverage without enforcing specific thresholds. To enable this functionality in your GitHub workflow, include the `continue-on-error: true` property in the job step configuration. This ensures that the workflow proceeds even if the coverage check fails.

Below is an example that reports files with coverage below 80% without causing the workflow to fail:
```yml
  - name: check test coverage
    id: coverage
    uses: vladopajic/go-test-coverage@v2
    continue-on-error: true
    with:
      profile: cover.out
      threshold-file: 80
```

## Report Coverage Difference

Using go-test-coverage, you can display a detailed comparison of code coverage changes relative to the base branch. When this feature is enabled, the report highlights files with coverage differences compared to the base branch.

The same logic is used in workflow in [this repo](/.github/workflows/test.yml). 
Example of report that includes coverage difference is [this PR](https://github.com/vladopajic/go-test-coverage/pull/129).

```yml
  # Download main (aka base) branch breakdown
  - name: download artifact (main.breakdown)
    id: download-main-breakdown
    uses: dawidd6/action-download-artifact@v6
    with:
      branch: main
      workflow_conclusion: success
      name: main.breakdown
      if_no_artifact_found: warn

  - name: check test coverage
    uses: vladopajic/go-test-coverage@v2
    with:
      config: ./.github/.testcoverage.yml
      profile: ubuntu-latest-profile,macos-latest-profile,windows-latest-profile

      # Save current coverage breakdown if current branch is main. It will be  
      # uploaded as artifact in step below.
      breakdown-file-name: ${{ github.ref_name == 'main' && 'main.breakdown' || '' }}

      # If this is not main brach we want to show report including
      # file coverage difference from main branch.
      diff-base-breakdown-file-name: ${{ steps.download-main-breakdown.outputs.found_artifact == 'true' && 'main.breakdown' || '' }}
    
  - name: upload artifact (main.breakdown)
    uses: actions/upload-artifact@v4
    if: github.ref_name == 'main'
    with:
      name: main.breakdown
      path: main.breakdown # as specified via `breakdown-file-name`
      if-no-files-found: error
```

## Post Coverage Report to PR

Here is an example of how to post comments with the coverage report to your pull request. 

The same logic is used in workflow in [this repo](/.github/workflows/test.yml).
Example of report is in [this PR](https://github.com/vladopajic/go-test-coverage/pull/129).

```yml
  - name: check test coverage
    id: coverage
    uses: vladopajic/go-test-coverage@v2
    continue-on-error: true # Should fail after coverage comment is posted
    with:
      config: ./.github/.testcoverage.yml
    
    # Post coverage report as comment (in 3 steps)
  - name: find pull request ID
    run: |
        PR_DATA=$(curl -s -H "Authorization: token ${{ secrets.GITHUB_TOKEN }}" \
        "https://api.github.com/repos/${{ github.repository }}/pulls?head=${{ github.repository_owner }}:${{ github.ref_name }}&state=open")
        PR_ID=$(echo "$PR_DATA" | jq -r '.[0].number')
        
        if [ "$PR_ID" != "null" ]; then
        echo "pull_request_id=$PR_ID" >> $GITHUB_ENV
        else
        echo "No open pull request found for this branch."
        fi
  - name: find if coverage report is already present
    if: env.pull_request_id
    uses: peter-evans/find-comment@v3
    id: fc
    with:
      issue-number: ${{ env.pull_request_id }}
      comment-author: 'github-actions[bot]'
      body-includes: 'go-test-coverage report:'
  - name: post coverage report
    if: env.pull_request_id
    uses: peter-evans/create-or-update-comment@v4
    with:
      token: ${{ secrets.GITHUB_TOKEN }}
      issue-number: ${{ env.pull_request_id }}
      comment-id: ${{ steps.fc.outputs.comment-id }}
      edit-mode: replace
      body: |
        go-test-coverage report:
        ```
        ${{ fromJSON(steps.coverage.outputs.report) }} ```

  - name: "finally check coverage"
    if: steps.coverage.outcome == 'failure'
    shell: bash
    run: echo "coverage check failed" && exit 1
```

## Generate Coverage Badge

Instructions for badge creation are available [here](./badge.md).
