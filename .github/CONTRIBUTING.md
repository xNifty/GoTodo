# Contributing

Thanks for contributing to this project.

## Before you start

- Check existing issues and pull requests to avoid duplicate work.
- Open an issue or discuss larger changes before spending significant time on implementation.
- Keep pull requests focused on one change whenever possible.
- Do not include unrelated formatting changes, generated files, or drive-by refactors.

## Branching policy

The repository uses the following branch model:

| Branch type | Create from | Pull request target | Purpose |
|---|---|---|---|
| `feat/*` | `dev` | `dev` | New functionality |
| `fix/*` | `dev` | `dev` | Non-urgent bug fixes |
| `chore/*` | `dev` | `dev` | Maintenance, dependencies, tooling |
| `docs/*` | `dev` | `dev` | Documentation updates |
| `release/*` | `dev` | `main` | Release stabilization; maintainers only |
| `hotfix/*` | `main` | `main` | Urgent production fixes; maintainers only |

Examples:

```text
feat/user-notification-settings
fix/invalid-token-error
chore/update-dotnet-sdk
docs/contribution-guidelines
release/1.8.0
hotfix/1.8.1
```

## Pull request rules

- Normal changes must target `dev`.
- Only authorized maintainers may open `release/*` or `hotfix/*` pull requests targeting `main`.
- External contributors should work from a fork and target `dev`.
- A PR opened against the wrong target may be asked to retarget and rebase onto current `dev` before review.
- PRs must pass required checks and address review feedback before merging.
- Do not force-push to protected branches.

## Before opening a PR

1. Rebase or merge the latest `dev` into your branch.
2. Run the relevant build, tests, formatter, and linter locally.
3. Update tests for behavior changes.
4. Update documentation, configuration examples, or changelog entries where relevant.
5. Write a clear PR description explaining what changed and why.

## Release process

1. A maintainer creates `release/x.y.z` from `dev`.
2. The release branch accepts only release-specific fixes and versioning work.
3. The release branch is merged into `main`.
4. The shipped `main` commit receives a version tag such as `v1.8.0`.
5. The release branch is merged back into `dev`.
6. The release branch is deleted.

## Hotfix process

1. A maintainer creates `hotfix/x.y.z` from `main`.
2. The hotfix is reviewed and merged into `main`.
3. The shipped commit is tagged and released.
4. The same fix is merged or cherry-picked back into `dev`.
5. If a release branch is active, apply the fix there as well.

## Code expectations

- Keep changes small, readable, and testable.
- Prefer clear naming over clever abstractions.
- Add or update tests when changing behavior.
- Do not commit secrets, credentials, local configuration, or private keys.
- Include migration notes when changes affect setup, configuration, APIs, or deployment.
