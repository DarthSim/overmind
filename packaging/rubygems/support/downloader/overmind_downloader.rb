# frozen_string_literal: true

require_relative "base_downloader"
require_relative "../../lib/overmind/version"

module Overmind
  module Downloader
    # Class for downloading Overmind binaries.
    class OvermindDownloader < BaseDownloader
      NAME = "overmind"
      DEFAULT_VERSION = ::Overmind::VERSION
      BASE_URL = "https://github.com/DarthSim/overmind/releases/download/v%s/%s"
      TARGET_PATH = "libexec/overmind"
      ALLOWED_PLATFORMS = %w[
        linux-arm linux-arm64 linux-386 linux-amd64 macos-amd64 macos-arm64 freebsd-386 freebsd-amd64 freebsd-arm
      ].freeze
      FILE_FORMAT = "%s-v%s-%s-%s.gz"

      private

      def extract_file(tmp_path)
        extracted_path = tmp_path.sub(".gz", "")

        system("gunzip -f #{tmp_path}")

        FileUtils.rm(target_path) if File.exist?(target_path)
        FileUtils.mkdir_p(File.dirname(target_path))
        FileUtils.mv(extracted_path, target_path)
        FileUtils.chmod("+x", target_path)
      end
    end
  end
end
