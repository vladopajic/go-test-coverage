# go-test-coverage GitHub Action


# Usage

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

# Action Outputs

The GitHub Action will set the following outputs, which can be used later in your workflow:

| Name            | Description                  |
|-----------------|------------------------------|
|`total-coverage` | Integer value in the range [0-100], representing the overall project test coverage percentage. |
|`badge-color`    | Color hex code for the badge  (e.g., `#44cc11`), representing the coverage status. |
|`badge-text`     | Deprecated! Text label for the badge. |
|`report`         | JSON-encoded string containing the detailed test coverage report. |

Note: Action outputs and inputs are also documented in [action.yml](/action.yml) file.

# Post Coverage Report to PR

Here is an example of how to post comments with the coverage report to your pull request. The same logic is used in workflow in [this repo](/.github/workflows/test.yml).

```yml
  - name: check test coverage
    id: coverage
    uses: vladopajic/go-test-coverage@v2
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
        body: |
        go-test-coverage report:
        ```
        ${{ fromJSON(steps.coverage.outputs.report) }} 
        ```
        edit-mode: replace
```          