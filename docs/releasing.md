# Releasing Fe

Use this checklist when preparing and publishing a stable CLI release.

## Before publishing

- [ ] CI passes on the intended release commit.
- [ ] Review the changelog and version number.
- [ ] Check the README examples and CLI help for accuracy.
- [ ] Confirm the release workflow builds the supported platform binaries and packages.
- [ ] Confirm the checksum manifest is generated and included with the release.

## After publishing

- [ ] Verify the GitHub Release has the expected tag and release notes.
- [ ] Check that the expected binaries, packages, and checksum manifest are attached.
- [ ] Download and smoke-test a release binary on a clean environment.
- [ ] Install and run a Linux package where applicable.
- [ ] Confirm the rolling release remains separate from the stable release.
- [ ] Check the published website and its download links.
