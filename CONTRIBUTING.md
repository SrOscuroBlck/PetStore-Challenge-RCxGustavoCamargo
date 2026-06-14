# Contributing & Workflow

The development flow for this repository. It follows GitHub Flow: `main` is always
stable and deployable, and all work happens on short-lived branches merged via pull request.

## Branching

Branch off `main`, named for the work and its issue:

```
feature/issue-<n>-<short-slug>     # new capability
bugfix/issue-<n>-<short-slug>      # fix
chore/issue-<n>-<short-slug>       # tooling, infra, maintenance
```

## Commits

[Conventional Commits](https://www.conventionalcommits.org/), referencing the issue:

```
feat: add merchant create-pet mutation (#12)
fix: prevent double-sell under concurrent checkout (#23)
chore: wire golangci-lint into CI (#2)
```

## Pull requests

- One PR per issue; the PR description links the issue with `Closes #<n>`.
- CI (build, vet, lint, race tests) must be green before merge.
- Squash-merge to keep `main` history linear.

## Versioning & releases

- [Semantic Versioning](https://semver.org). Pre-1.0 milestones ship as `v0.x.0`;
  `v1.0.0` is the submission-ready release.
- A release is cut by tagging `main`: `git tag vX.Y.Z && git push origin vX.Y.Z`.
- The tag triggers the release workflow (build, GitHub Release with notes, and a
  container image once a Dockerfile exists).
- Update the `[Unreleased]` section of `CHANGELOG.md` in each PR.
