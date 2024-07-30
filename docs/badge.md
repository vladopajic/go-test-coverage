# Coverage badge

Repositories which use `go-test-coverage` action in their workflows could easily create beautiful coverage badge and embed them in markdown files (eg. ![coverage](https://raw.githubusercontent.com/vladopajic/go-test-coverage/badges/.badges/main/coverage.svg)).


`go-test-coverage` offers multiple mechanisms for creating and storing these badges:
* generate a badge and save it within the same GitHub repository (ideal for public repositories)
* generate a badge and store it on a Content Delivery Network (CDN) (ideal for private repositories)
* generate a badge and choose a custom method for storage, allowing flexibility to align with your repository's specific requirements
* generate a badge and store it in another public GitHub repository

## Coverage badge hosted on same GitHub repository

`go-test-coverage` can create a coverage badge and automatically commit it to the same GitHub repository. This feature is particularly well-suited for public repositories.

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
- Allow some time for GitHub to do it's thing if file is not immediately accessible via it's link.
- `check test coverage` step may fail if GitHub token does not have permissions to write. To fix this you can:
  - set write permission to job with `permissions: write-all` directive, or
  - give write permissions to GitHub token: Go to repository Settings -> Actions -> Workflow Permissions section and give actions Read and Write permissions
- For private repositories this will not work because only content from public repository could be accessible via `raw.githubusercontent.com`. For private repositories coverage badge could be hosted with methods described below.

## Coverage badge hosted on CDN

`go-test-coverage` can generate a coverage badge and upload it to a content delivery network (CDN) such as Amazon S3 or DigitalOcean Spaces. This option is especially suitable for private repositories.

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

## Generate badge

With `go-test-coverage`, you can generate a coverage badge and save it locally on the file system. This badge file can be managed and utilized using a custom mechanism tailored to your specific needs.

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

## Coverage badge hosted on another public GitHub repository

Just like the method where badges are hosted within the same repository, `go-test-coverage` provides a straightforward configuration to commit badges to any other repository, preferably one that is public. 

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

Like in the first example `badges` branch should be created with same method as described above.

## Badge examples

Badge examples created with this method would look like this:

![coverage](https://raw.githubusercontent.com/vladopajic/go-test-coverage/badges/.badges/badge-examples/coverage-0.svg)
![coverage](https://raw.githubusercontent.com/vladopajic/go-test-coverage/badges/.badges/badge-examples/coverage-50.svg)
![coverage](https://raw.githubusercontent.com/vladopajic/go-test-coverage/badges/.badges/badge-examples/coverage-70.svg)
![coverage](https://raw.githubusercontent.com/vladopajic/go-test-coverage/badges/.badges/badge-examples/coverage-80.svg)
![coverage](https://raw.githubusercontent.com/vladopajic/go-test-coverage/badges/.badges/badge-examples/coverage-90.svg)
![coverage](https://raw.githubusercontent.com/vladopajic/go-test-coverage/badges/.badges/badge-examples/coverage-100.svg)
