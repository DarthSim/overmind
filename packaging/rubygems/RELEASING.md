## Prepare gems

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
