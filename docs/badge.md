# Coverage Badge

Repositories which use `go-test-coverage` action in their workflows could easily create beautiful coverage badge and embed them in markdown files (eg. ![coverage](https://raw.githubusercontent.com/vladopajic/go-test-coverage/badges/.badges/main/coverage.svg)).

## Coverage Badge hosted on GitHub repository

`go-test-coverage` can generate coverage badge and commit it to designated GitHub repository. 

Example:
```yml
- name: check test coverage
  uses: vladopajic/go-test-coverage@v2
  with:
    profile: cover.out
    local-prefix: github.com/org/project
    threshold-total: 95

    ## name of branch where badges are stored
    ## ideally this should be orphan branch, see below how to create this branch
    git-branch: badges 
    ## git-token is needed to push to repository
    ## when token is not specified (value is '') this feature is turend off
    ## in this example badge is created and committed only for main brach
    git-token: ${{ github.ref_name == 'main' && secrets.GITHUB_TOKEN || '' }}
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

Lastly, check output of `check test coverage` step to see markdown snippet which can be added to markdown files. 

If instruction from this example was followed through, this link should be

```markdown
![coverage](https://raw.githubusercontent.com/org/project/badges/.badges/main/coverage.svg)
```

where `org/project` part would match corresponding project.

Notes:
- Allow some time for Github to do it's thing if file is not immediately accessible via this link.
- `check test coverage` step may fail if GitHub token does not have permissions to write. To fix this you can:
  - add permission to job with `permissions: write-all` directive, or
  - give permissions to GitHub token: Go to repository Settings -> Actions -> Workflow Permissions section and give actions Read and Write permissions
- For private repositories this will not work because only content from public repository could be accessible via `raw.githubusercontent.com`. For private repositories coverage badge could be hosted via CDN, as described below.

## Coverage Badge hosted on CDN

`go-test-coverage` can generate coverage badge and upload it to CDN like Amazon S3 or DigitalOcean Spaces. 

Example:
```yml
- name: check test coverage
  uses: vladopajic/go-test-coverage@v2
  with:
    profile: cover.out
    local-prefix: github.com/org/project
    threshold-total: 95

    ## when secret is not specified (value is '') this feature is turend off
    ## in this example badge is created and uploaded only for main brach
    cdn-secret:  ${{ github.ref_name == 'main' && secrets.CDN_SECRET || '' }}
    cdn-key: ${{ secrets.CDN_KEY }}
    ## in case of DigitalOcean Spaces use `us-ease-1` always, otherwise use region of your CDN
    cdn-region: us-east-1 
    ## in case of DigitalOcean Spaces endpoint should be set with region and without bucket
    cdn-endpoint: https://nyc3.digitaloceanspaces.com 
    cdn-file-name: .badges/${{ github.ref_name }}/coverage.svg
    cdn-bucket-name: my-bucket-name
    cdn-force-path-style: false
```

## Badge examples

Badge examples created with this method would look like this:

![coverage](https://raw.githubusercontent.com/vladopajic/go-test-coverage/badges/.badges/badge-examples/coverage-0.svg)
![coverage](https://raw.githubusercontent.com/vladopajic/go-test-coverage/badges/.badges/badge-examples/coverage-50.svg)
![coverage](https://raw.githubusercontent.com/vladopajic/go-test-coverage/badges/.badges/badge-examples/coverage-70.svg)
![coverage](https://raw.githubusercontent.com/vladopajic/go-test-coverage/badges/.badges/badge-examples/coverage-80.svg)
![coverage](https://raw.githubusercontent.com/vladopajic/go-test-coverage/badges/.badges/badge-examples/coverage-90.svg)
![coverage](https://raw.githubusercontent.com/vladopajic/go-test-coverage/badges/.badges/badge-examples/coverage-100.svg)
