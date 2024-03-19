## Releasing

### Requirements
It assumes that there are already `overmind` and `tmux` binaries in the Github [release](https://github.com/DarthSim/overmind/releases/tag/v2.4.0).

## Build gem

To build gems for specific platforms, your Ruby version must be at least 3.0.
This requirement exists because Ruby 3.0 introduced enhancements in RubyGems, including support for the `--platform` argument,
enabling the specification of target platforms for gem files.

For further details on this enhancement, refer to the [RubyGems changelog](https://github.com/rubygems/rubygems/blob/master/CHANGELOG.md#enhancements-91).

First, run bundle install command

```bash
bundle install
```

Generate gem files for various platforms with:

```bash
rake "overmind:build:all"
```

This command builds gems for Linux, macOS, and FreeBSD.

To build a gem for a specific platform:

```bash
rake "overmind:build:linux[amd64]" # linux x86_64
rake "overmind:build:macos[arm64]" # macos arm64
```

## Prepare tmux
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
