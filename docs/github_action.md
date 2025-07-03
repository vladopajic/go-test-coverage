# go-test-coverage GitHub Action

The `go-test-coverage` GitHub Action provides the following capabilities:
- Enforce success of GitHub workflows only when a specified coverage threshold is met.
- Generate a coverage badge to display the total test coverage.
- Post a detailed coverage report as a comment on pull requests, including:
  - current test coverage
  - uncovered lines (reported when any threshold is not satisfied)
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
    threshold-file: 80
    threshold-package: 80
    threshold-total: 95
```

Note: When using a config file alongside action properties, specifying these parameters will override the corresponding values in the config file.

## Generate Coverage Badge

Instructions for badge creation are available [here](./badge.md).

## Source Directory

Some projects, such as monorepos with multiple projects under the root directory, may require specifying the path to a project's source.
In such cases, the `source-dir` property can be used to specify the source files location relative to the root directory.

```yml
- name: check test coverage
  uses: vladopajic/go-test-coverage@v2
  with:
    config: ./.testcoverage.yml
    source-dir: ./some_project
```

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

# Post coverage report as comment (in 2 steps)
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
- name: post coverage report
  if: env.pull_request_id
  uses: thollander/actions-comment-pull-request@v3
  with:
    github-token: ${{ secrets.GITHUB_TOKEN }}
    comment-tag: coverage-report
    pr-number: ${{ env.pull_request_id }}
    message: |
      go-test-coverage report:
      ```
      ${{ fromJSON(steps.coverage.outputs.report) }}```

- name: "finally check coverage"
  if: steps.coverage.outcome == 'failure'
  shell: bash
  run: echo "coverage check failed" && exit 1
```


---


# Advanced

## Types of GitHub Actions

The `go-test-coverage` project provides two types of GitHub Actions:

- **Binary-based Action (default)**
  
  Executes the compiled binary using a Docker image. This is the default action, defined in [/action.yml](/action.yml).
  
  Usage, as demonstrated throughout the documentation:
  ```yml
  - name: check test coverage
    uses: vladopajic/go-test-coverage@v2
    with: ...
  ```
  
- **Source-based Action (optional/experimental)**
  
  Runs the source code using the `go run` command. This experimental action is defined in [/action/source/action.yml](/action/source/action.yml).

  Usage:
  ```yml
  - name: check test coverage
    # note: uses property adds 'action/source' part, compared to default action
    uses: vladopajic/go-test-coverage/action/source@v2
    with: ...
  ```
   Note: this action requires `go` to be installed.

Both actions have the same inputs, so they can be used interchangeably.




 
