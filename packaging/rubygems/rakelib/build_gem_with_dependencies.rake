# frozen_string_literal: true

require_relative "../support/downloader/overmind_downloader"
require_relative "../support/downloader/tmux_downloader"

GEM_PLATFORMS = {
  "linux-arm" => "arm-linux",
  "linux-arm64" => "arm64-linux",
  "linux-386" => "x86-linux",
  "linux-amd64" => "x86_64-linux",
  "macos-amd64" => "x86_64-darwin",
  "macos-arm64" => "arm64-darwin",
  "freebsd-386" => "x86-freebsd",
  "freebsd-amd64" => "x86_64-freebsd",
  "freebsd-arm" => "arm-freebsd"
}.freeze

namespace :overmind do
  namespace :build do
    desc "Download `Overmind` and `tmux` binaries, and prepare gem"
    task :all do
      GEM_PLATFORMS.each do |file_platform, gem_platform|
        os, arch = file_platform.split("-")

        Overmind::Downloader::OvermindDownloader.new(os: os, arch: arch).call
        Overmind::Downloader::TmuxDownloader.new(os: os, arch: arch).call

        system("gem build overmind.gemspec --platform #{gem_platform}")
      end
    end

    desc "Prepare gem for Linux with optional ARCH (e.g., rake overmind:build:linux[arm64])"
    task :linux, [:arch] do |_t, args|
      build_for_os("linux", args[:arch])
    end

    desc "Prepare gem for macOS with optional ARCH (e.g., rake overmind:build:macos[amd64])"
    task :macos, [:arch] do |_t, args|
      build_for_os("macos", args[:arch])
    end

    desc "Prepare gem for FreeBSD with optional ARCH (e.g., rake overmind:build:freebsd[amd64])"
    task :freebsd, [:arch] do |_t, args|
      build_for_os("freebsd", args[:arch])
    end

    def build_for_os(os, arch = nil)
      platforms = GEM_PLATFORMS.select { |k, _| k.start_with?(os) }
      platforms = platforms.select { |k, _| k.end_with?(arch) } if arch

      platforms.each do |file_platform, gem_platform|
        os, arch = file_platform.split("-")

        Overmind::Downloader::OvermindDownloader.new(os: os, arch: arch).call
        Overmind::Downloader::TmuxDownloader.new(os: os, arch: arch).call

        system("gem build overmind.gemspec --platform #{gem_platform}")
      end
    end
  end
end
