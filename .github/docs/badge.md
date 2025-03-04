# Coverage badge

Repositories using the go-test-coverage GitHub action can easily generate and embed coverage badges in markdown files, allowing you to visualize and track your test coverage. For example: 

![coverage](https://raw.githubusercontent.com/vladopajic/go-test-coverage/badges/.badges/main/coverage.svg)


The go-test-coverage action supports several methods for generating and storing these badges, depending on your repository’s needs:
- **Storing badges within the same GitHub repository** (ideal for public repositories).
- **Storing badges on a Content Delivery Network (CDN)** (suitable for private repositories).
- **Using a custom method for storage**, allowing flexibility.
- **Storing badges in a different public GitHub repository**.

## Hosting the Coverage Badge in the Same GitHub Repository

For public repositories, `go-test-coverage` can automatically create and commit a badge to the same GitHub repository. This is especially useful for keeping all assets together within the project.

Example:
```yml
- name: check test coverage
  uses: vladopajic/go-test-coverage@v2
  with:
    profile: cover.out
    local-prefix: github.com/org/project
    threshold-total: 95

    ## when token is not specified (value '') this feature is turned off
    ## in this example badge is created and committed only for main branch
    git-token: ${{ github.ref_name == 'main' && secrets.GITHUB_TOKEN || '' }}
    ## name of branch where badges are stored
    ## ideally this should be orphan branch (see below how to create this branch)
    git-branch: badges 
```

Orphan branch (has no history from other branches) needs to be created prior to running this workflow, to create an orphan branch manually:

```bash
git checkout --orphan badges
git rm -rf .
rm -f .gitignore
echo '# Badges' > README.md
git add README.md
git commit -m 'init'
git push origin badges
```

Once the workflow completes, check the output of the `Check test coverage` step for a markdown snippet to embed the badge in your documentation. The link should look like:

```markdown
![coverage](https://raw.githubusercontent.com/org/project/badges/.badges/main/coverage.svg)
```

Notes:
- Allow some time for GitHub to make the file available via its link.
- The workflow may fail if the GitHub token doesn’t have write permissions. To fix this:
  - Set `permissions: write-all` in your job configuration, or
  - Navigate to repository settings → Actions → Workflow permissions, and grant Read and Write permissions.
- This method only works for public repositories because private repository content is not accessible via `raw.githubusercontent.com`. For private repositories, refer to the CDN method described below.

## Hosting the Coverage Badge on a CDN

For private repositories, `go-test-coverage` can generate a badge and upload it to a CDN like Amazon S3 or DigitalOcean Spaces, making it accessible while keeping the repository private.

Example:
```yml
- name: check test coverage
  uses: vladopajic/go-test-coverage@v2
  with:
    profile: cover.out
    local-prefix: github.com/org/project
    threshold-total: 95

    ## when secret is not specified (value '') this feature is turned off.
    ## in this example badge is created and uploaded only for main branch.
    cdn-secret:  ${{ github.ref_name == 'main' && secrets.CDN_SECRET || '' }}
    cdn-key: ${{ secrets.CDN_KEY }}
    ## in case of DigitalOcean Spaces use `us-ease-1` always as region,
    ## otherwise use region of your CDN.
    cdn-region: us-east-1 
    ## in case of DigitalOcean Spaces endpoint should be with region and without bucket
    cdn-endpoint: https://nyc3.digitaloceanspaces.com 
    cdn-file-name: .badges/${{ github.repository }}/${{ github.ref_name }}/coverage.svg
    cdn-bucket-name: my-bucket-name
    cdn-force-path-style: false
```

## Generating a Local Badge

`go-test-coverage` can also generate a badge and store it locally on the file system, giving you the flexibility to handle badge storage through custom methods.

Example:
```yml
- name: check test coverage
  uses: vladopajic/go-test-coverage@v2
  with:
    profile: cover.out
    local-prefix: github.com/org/project
    threshold-total: 95

    # badge will be generated and store on file system with `coverage.svg` name
    badge-file-name: coverage.svg

  ## ... implement your method for storing badge 
```

## Hosting the Badge in Another Public GitHub Repository

You can also store the coverage badge in a separate public GitHub repository, which is particularly useful when managing multiple repositories or projects.

Example:
```yml
- name: check test coverage
  uses: vladopajic/go-test-coverage@v2
  with:
    profile: cover.out
    local-prefix: github.com/org/project
    threshold-total: 95

    ## in this case token should be from other repository that will host badges.
    ## this token is provided via secret `BADGES_GITHUB_TOKEN`.
    git-token: ${{ github.ref_name == 'main' && secrets.BADGES_GITHUB_TOKEN || '' }}
    git-branch: badges
    ## repository should match other repository where badges are hosted.
    ## format should be `{owner}/{repository}`
    git-repository: org/badges-repository
    ## use custom file name that will have repository name as prefix.
    ## this could be useful if you want to create badges for multiple repositories.
    git-file-name: .badges/${{ github.repository }}/${{ github.ref_name }}/coverage.svg
```

Ensure the `badges` branch is created in the target repository using the same steps as described for orphan branches earlier.

## Badge Examples

Here are some example badges generated with this method:

![coverage](https://raw.githubusercontent.com/vladopajic/go-test-coverage/badges/.badges/badge-examples/coverage-0.svg)
![coverage](https://raw.githubusercontent.com/vladopajic/go-test-coverage/badges/.badges/badge-examples/coverage-50.svg)
![coverage](https://raw.githubusercontent.com/vladopajic/go-test-coverage/badges/.badges/badge-examples/coverage-70.svg)
![coverage](https://raw.githubusercontent.com/vladopajic/go-test-coverage/badges/.badges/badge-examples/coverage-80.svg)
![coverage](https://raw.githubusercontent.com/vladopajic/go-test-coverage/badges/.badges/badge-examples/coverage-90.svg)
![coverage](https://raw.githubusercontent.com/vladopajic/go-test-coverage/badges/.badges/badge-examples/coverage-100.svg)
