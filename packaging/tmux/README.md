## Compile tmux with dependencies

We can build tmux binaries for Linux platforms (amd64/arm64/arm/386) via `Docker buildx' using
```bash
docker buildx build --platform linux/amd64,linux/arm64,linux/arm/v7,linux/386 -t tmux-multiarch:latest --output type=local,dest=./dist data
```

Run into `segfaults`? See this [GitHub issue for solutions] (https://github.com/docker/buildx/issues/314#issuecomment-1043156006).


For MacOS, the `tmux` binaries can be prepared by running `scripts/generate_tmux.sh` on a MacOS machine:
```bash
./scripts/generate_tmux.sh
```

This script will also work on Linux, generating `tmux` binaries for that platform.
