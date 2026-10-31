# Updating Fe

Fe checks GitHub Releases for the latest stable or rolling build, then delegates stable updates to the package manager that owns the running installation.

```sh
fe update check
fe update --dry-run
fe update
fe update --channel rolling
```

## Safety rules

- Installation detection uses package-manager records and verifies that the running executable belongs to the recorded package.
- Fe does not infer a direct-binary installation from the executable directory.
- An explicit `--method` still has to pass package-ownership verification; it does not bypass detection.
- Fe checks the package manager's available version against the latest GitHub stable release. If the configured source is stale or has no candidate, Fe refuses to run an update that cannot reach the current release.
- Fe prints the exact delegated command and asks before running it. `--dry-run` prints the command without running it.
- Fe never adds or changes package-manager sources.
- Package managers do not currently publish Fe's rolling channel. Rolling checks are available, but applying rolling updates is not enabled yet.
- Direct-binary replacement, checksum-verified installation, rollback, and delta updates are not part of this stage.

## Supported package-manager records

| Method | Ownership check | Update command |
| --- | --- | --- |
| Chocolatey | Installed Fe nuspec, payload, and matching Chocolatey shim | `choco upgrade fe -y --no-progress` |
| Homebrew | Installed formula metadata and running executable listed by the formula | `brew upgrade fe` |
| APT | dpkg ownership of the running executable | `sudo apt-get install --only-upgrade fe` |

APT delegation only works when an already-configured APT source offers the current Fe release. Fe publishes standalone Debian packages, not an APT repository. The Homebrew tap and Chocolatey feed must also have a package at least as new as the latest GitHub release; otherwise Fe reports that the source is behind and stops.

## Installation method override

Use `--method` when automatic detection finds multiple matching package records:

```sh
fe update check --method chocolatey
fe update --method homebrew --dry-run
```

An override selects which metadata record to inspect. It does not force an update through an unrelated package manager. Direct binaries are reported as unsupported until the separate verified updater is implemented.
