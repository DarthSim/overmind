# frozen_string_literal: true

require_relative "base_downloader"

module Overmind
  module Downloader
    # Class for downloading Tmux binaries.
    class TmuxDownloader < BaseDownloader
      NAME = "tmux"
      DEFAULT_VERSION = "3.4"
      BASE_URL = "https://github.com/DarthSim/overmind/releases/download/v2.4.0/%s"
      TARGET_PATH = "libexec/prebuilt-tmux"
      ALLOWED_PLATFORMS = %w[
        linux-arm linux-arm64 linux-386 linux-amd64 macos-amd64 macos-arm64
      ].freeze

      FILE_FORMAT = "%s-v%s-%s-%s.tar.gz"

      private

      # Extracts the file to the specified target path.
      def extract_file(tmp_path)
        FileUtils.rm_rf(target_path) if File.exist?("#{target_path}/bin/tmux")
        FileUtils.mkdir_p(target_path)

        system("tar -xzf #{tmp_path} -C #{target_path}")

        tmux_executable_path = File.join(target_path, "bin/tmux")
        FileUtils.chmod("+x", tmux_executable_path) if File.exist?(tmux_executable_path)
      end

      def prepare_download_url
        format(@base_url, file_name)
      end
    end
  end
end
