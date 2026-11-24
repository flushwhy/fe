# Getting started with Fe

Fe is a local CLI that connects a project to the tools it already uses. It can scaffold starters, inspect a workspace, check toolchains, run configured builds, help with engine workflows, prepare artifacts, and publish releases.

Fe is not a compiler or game engine. Your project remains a normal, rebuildable project that can still be built with its native tools.

## 1. Install Fe

Use a stable release for day-to-day work. Rolling builds contain the newest development changes and can change without notice.

### Windows

Download the Windows executable from [GitHub Releases](https://github.com/flushwhy/fe/releases). Put it in a directory on your `PATH`, then open a new terminal.

If the Fe package is available from your configured Chocolatey source, you can also install it with:

```powershell
choco install fe
```

### macOS

Homebrew:

```sh
brew tap flushwhy/tap
brew install fe
```

Or download the matching macOS binary from [GitHub Releases](https://github.com/flushwhy/fe/releases).

### Debian and Ubuntu

Download the matching amd64 or arm64 `.deb` package, then install it:

```sh
sudo apt install ./fe_VERSION_amd64.deb
```

### Other Linux distributions and FreeBSD

Download the matching standalone binary or archive from [GitHub Releases](https://github.com/flushwhy/fe/releases), make the binary executable, and put it somewhere on your `PATH`.

For example:

```sh
chmod +x fe-linux-amd64
sudo install -m 0755 fe-linux-amd64 /usr/local/bin/fe
```

### Build from source

Install Go, then run:

```sh
go install github.com/flushwhy/fe@latest
```

### Verify the download

Release assets include a `SHA256SUMS` file. Download it alongside your binary or package and verify the matching file before installing.

Linux:

```sh
sha256sum --check SHA256SUMS
```

macOS:

```sh
shasum -a 256 -c SHA256SUMS
```

Windows PowerShell:

```powershell
Get-FileHash .\\fe-windows-amd64.exe -Algorithm SHA256
```

Compare the displayed hash with the Windows executable entry in `SHA256SUMS`.

## 2. Create or inspect a project

Start from a language starter:

```sh
fe init go --name my-app
cd my-app
```

Or use a named template:

```sh
fe template list
fe init --template cpp-raylib-xmake --name starfall
cd starfall
```

Fe downloads named templates from [fe-templates](https://github.com/flushwhy/fe-templates) and caches them locally.

If you already have a project, Fe does not require you to recreate it. Run:

```sh
fe inspect
```

This is a read-only scan for project and ecosystem markers. It does not execute project code or modify files.

## 3. Check the tools you need

Use these commands for different levels of inspection:

| Command | What it does | Changes files? |
|---|---|---|
| `fe inspect` | Detects project types and shows informational build suggestions. | No |
| `fe setup` | Lists required and optional tools for the detected project. | No |
| `fe doctor` | Checks local tools and project configuration. | No |
| `fe preflight` | Checks detected runtimes and runs configured pre-build checks. | Runs configured checks |
| `fe check` | Runs available language-aware static analyzers. | Depends on the analyzer |

**Important:** `fe setup` is an inspection command. It does not install compilers, SDKs, engines, or other software.

`fe sync` is different: it synchronizes Fe-managed project files and required local tools. It is intended for developer machines and is disabled in CI. For example, it can prepare a project-managed Godot version when the project requests one and no matching global engine is found.

```sh
fe setup
fe doctor
fe preflight
```

Install any missing prerequisites using their official installers, then rerun the checks.

## 4. Configure a rebuildable project

Fe reads an optional `.fe.yaml` from the project directory. Keep the actual build command explicit and use the project's normal compiler or build system.

Example:

```yaml
build:
  targets:
    - name: linux-release
      platform: linux-x64
      cmd: "cmake --build build --config Release"

butler:
  directory: "builds"
```

The command above assumes you have already configured the CMake build directory. Fe runs the command you configure; it does not generate your CMake project or replace CMake.

Configuration priority is:

1. Environment variables
2. `.fe.yaml`
3. Fe conventions
4. Defaults

Keep generated build output separate from source files. A typical layout might look like:

```text
my-app/
├── .fe.yaml
├── CMakeLists.txt
├── src/
├── assets/
├── build/       # native build-system files
└── builds/      # distributable output
```

Commit the project files needed to rebuild the project: source, assets, build-system definitions, lockfiles, and the Fe configuration you intentionally maintain. Do not commit machine-specific SDK paths or local caches.

## 5. Build and iterate

Run the configured build targets:

```sh
fe build
fe build --list-targets
fe build --target linux-release
```

Other useful workflows include:

- `fe watch` — watch supported project files and rebuild supported C++ changes.
- `fe add clangd` — write or update editor/LSP configuration for clangd.
- `fe fix compile-commands` — repair a C/C++ compilation database for clangd.
- `fe godot` — inspect and work with Godot engines, exports, and native extensions.
- `fe o3de` — discover and manage O3DE engines, projects, Gems, and builds.
- `fe pack` — pack PNG sprites into spritesheets.
- `fe transcode` — convert audio or video through FFmpeg.

`fe add` configures editor or language-server files. It does not install the language server itself.

## 6. Package and publish

A typical release sequence is:

```sh
fe validate
fe preflight
fe build
fe manifest
fe package --format zip
fe release --dry-run
```

Review the dry-run output before publishing. The release command can run configured build and asset stages, package artifacts, and publish to itch.io through Butler when configured.

Useful publishing commands:

- `fe manifest` — write an artifact inventory with SHA-256 checksums and build metadata.
- `fe package` — create a ZIP or tar.gz archive from build artifacts.
- `fe changelog` — generate Markdown release notes from Git history and tags.
- `fe bmp` — upload configured build output to itch.io using Butler.
- `fe gumroad release` — upload a stable release to a pre-created, configured Gumroad product.

Install and authenticate external tools such as Butler, FFmpeg, or the Gumroad CLI separately. Fe does not silently install them.

## 7. Keep the project portable

For a project another developer or CI agent can rebuild:

- Document the compiler, engine, SDK, and external tools the project requires.
- Pin tool versions where the ecosystem supports it.
- Keep build commands and output locations in version-controlled project configuration.
- Keep secrets in environment variables or CI secret storage, never in `.fe.yaml`.
- Make sure the project builds with its native toolchain without relying on Fe-managed caches.
- Use `fe inspect`, `fe setup`, `fe doctor`, and `fe preflight` to explain and check the local environment.

Fe is a helper around your project—not a replacement for its source files or native build system.

## Updating Fe

Run `fe update check` to inspect the latest release and your detected installation method. Package-managed stable installations can be updated through their package manager when that source has caught up with GitHub Releases. Direct binaries can be checked but are not replaced by Fe yet. Rolling-channel checks are available, but rolling updates are not automated.

Use the release page to install a newer direct binary manually. Do not assume `fe update` can replace a binary installed outside a supported package manager.

## Command reference

Run `fe --help` for the full command tree, or `fe <command> --help` for flags and examples. The website's [command explorer](https://flushwhy.github.io/fe/#commands) provides a searchable overview.

## Help and issues

- [Fe source and issue tracker](https://github.com/flushwhy/fe)
- [Project templates](https://github.com/flushwhy/fe-templates)
- [GitHub Releases](https://github.com/flushwhy/fe/releases)
