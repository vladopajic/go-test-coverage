# Coverage Badge

Repositories which use `go-test-coverage` action in their workflows could easily create beautiful coverage badge and embed them in markdown files (eg. ![coverage](https://raw.githubusercontent.com/vladopajic/go-test-coverage/badges/.badges/main/coverage.svg)).

## Self hosted badges

Easiest way to add badge in markdown files is to generate them via workflow and self host them in git repository.

Generating self hosted coverage badge could be done with `action-badges/core` action. This action will create badge file and commit it to some orphan branch.

Workflow example:

```yml
name: Go test coverage check
runs-on: ubuntu-latest
steps:
  - uses: actions/checkout@v3
  - uses: actions/setup-go@v3
  
  - name: generate test generate coverage
    run: go test ./... -coverprofile=./cover.out

  - name: check test coverage
    id: coverage ## this step must have id
    uses: vladopajic/go-test-coverage@v2
    with:
      profile: cover.out
      local-prefix: github.com/org/project
      threshold-file: 80
      threshold-package: 80
      threshold-total: 95
  
  - name: make coverage badge
    uses: action-badges/core@0.2.2
    if: contains(github.ref, 'main')
    with:
      label: coverage
      message: ${{ steps.coverage.outputs.badge-text }}
      message-color: ${{ steps.coverage.outputs.badge-color }}
      file-name: coverage.svg
      badge-branch: badges ## orphan branch where badge will be committed
      github-token: "${{ secrets.GITHUB_TOKEN }}"
```

Orphan branch needs to be created prior to running this workflow, to create an orphan branch manually:

```bash
git checkout --orphan badges
git rm -rf .
rm -f .gitignore
echo '# Badges' > README.md
git add README.md
git commit -m 'init'
git push origin badges
```

Lastly, check output of `make coverage badge` step to see markdown snippet which can be added to markdown files. 

If instruction from this example was followed through, this link should be

```markdown
![coverage](https://raw.githubusercontent.com/org/project/badges/.badges/main/coverage.svg)
```

where `org/project` part would match corresponding project.

Allow some time for Github to do it's thing if file is not immediately accessible via this link

## Private repositories

In private repositories, files are not accessible without the inclusion of an access token in the URL. However, this practice of adding an access token directly to the URL has a drawback: contributors can potentially continue accessing the repository even after their permissions have been revoked. 

To mitigate this issue, consider incorporating an additional step into the workflow file that uploads the coverage badge file to a designated Content Delivery Network (CDN).


## Badge examples

Badge examples created with this method would look like this:

![coverage](https://raw.githubusercontent.com/vladopajic/go-test-coverage/badges/.badges/badge-examples/coverage-0.svg)
![coverage](https://raw.githubusercontent.com/vladopajic/go-test-coverage/badges/.badges/badge-examples/coverage-50.svg)
![coverage](https://raw.githubusercontent.com/vladopajic/go-test-coverage/badges/.badges/badge-examples/coverage-70.svg)
![coverage](https://raw.githubusercontent.com/vladopajic/go-test-coverage/badges/.badges/badge-examples/coverage-80.svg)
![coverage](https://raw.githubusercontent.com/vladopajic/go-test-coverage/badges/.badges/badge-examples/coverage-90.svg)
![coverage](https://raw.githubusercontent.com/vladopajic/go-test-coverage/badges/.badges/badge-examples/coverage-100.svg)
