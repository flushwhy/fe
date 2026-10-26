# Container builds

Fe can run a configured build command inside a temporary Docker or Podman container. Install and configure the runtime yourself; Fe does not install or manage it.

Configure an image and command in `.fe.yaml`:

```yaml
build:
  container:
    image: flushy/odin:nightly
    command: "odin build . -out:builds/game"
```

Run the build explicitly:

```sh
fe build --container
# Or, if you use Podman:
fe build --container --runtime podman
```

Fe mounts the current project directory at `/src` and runs the command there. Write output into the mounted project directory so the resulting files are available on the host. The container is removed after the command exits, including failed builds; the image remains cached by the container runtime. If the image is not present locally, the runtime will attempt to pull it.

Container builds are opt-in and separate from the normal `fe build` targets. They do not automatically cross-compile: the image must contain a compiler and any SDKs/toolchains needed for the requested target. The configured command runs through `sh -lc`, so use a Linux-compatible image with a shell.
