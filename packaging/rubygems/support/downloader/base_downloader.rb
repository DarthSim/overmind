# frozen_string_literal: true

require "faraday"
require "faraday/follow_redirects"
require "fileutils"

module Overmind
  module Downloader
    # Abstract base class for downloading files.
    # Provides common interfaces and utilities for subclasses.
    class BaseDownloader
      attr_reader :version, :os, :arch, :file_name, :download_url, :target_path, :uri

      # Initializes a new downloader instance.
      # @param os [String] the operating system for the download.
      # @param arch [String] the system architecture for the download.
      # @param version [String] the version of the file to download.
      def initialize(os:, arch:, version: self.class::DEFAULT_VERSION)
        @version = version
        @os = os.downcase
        @arch = arch.downcase
        @base_url = self.class::BASE_URL
        @target_path = self.class::TARGET_PATH
        @file_format = self.class::FILE_FORMAT

        @file_name = format(@file_format, self.class::NAME, version, os, arch)
        @download_url = prepare_download_url
      end

      def call
        unless allowed_platform?
          puts "Unsupported platform: #{"#{@os}-#{@arch}"}"

          FileUtils.rm_rf(target_path)

          return false
        end

        tmp_path = download_file
        extract_file(tmp_path)
      end

      private

      def allowed_platform?
        platform_key = "#{@os}-#{@arch}"

        self.class::ALLOWED_PLATFORMS.include?(platform_key)
      end

      def download_file
        tmp_path = File.join("tmp", file_name)
        response = faraday_builder.get(download_url)

        File.write(tmp_path, response.body, mode: "wb")

        tmp_path
      end

      def extract_file(_tmp_path)
        raise NoMethodError
      end

      def faraday_builder(args = {})
        Faraday.new(args) do |builder|
          builder.response :logger
          builder.response :follow_redirects
          builder.response :raise_error

          yield(builder) if block_given?
        end
      end

      def prepare_download_url
        format(@base_url, version, file_name)
      end
    end
  end
end
