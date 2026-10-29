# Releasing Fe

Use this checklist for a stable CLI release.

## Before merging the release PR

- [ ] GitHub Actions **Build and test** passes on the intended commit.
- [ ] Review the Release Please changelog and version bump; keep user-facing changes under the right headings.
- [ ] Check CLI help and README examples for commands and flags introduced since the previous release.
- [ ] Check the docs site command reference and installation instructions against the current release workflow.
- [ ] Confirm the stable release workflow still builds the supported platform binaries, Linux packages, and checksum manifest.

## After publishing

- [ ] Open the GitHub Release and verify the version tag and generated notes.
- [ ] Confirm expected binaries/packages and `SHA256SUMS` are attached.
- [ ] Download and smoke-test at least one binary on a clean environment; for Linux packages, install and run `fe --help`.
- [ ] Verify the rolling release remains separate from the stable release.
- [ ] Open the deployed site and check navigation, quick-start commands, and links on desktop and mobile.

Do not mark a release ready based only on a successful compile: CI, release assets, install smoke tests, and published docs are separate checks.
