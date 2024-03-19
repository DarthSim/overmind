# frozen_string_literal: true

module Overmind
  # Command-line interface for running golang library Overmind
  # It ensures that the necessary dependencies, such as tmux, are present
  # and runs Overmind with any provided arguments.
  class CLI
    SUPPORTED_OS_REGEX = /darwin|linux|bsd/i
    # Path to the library's executable files
    LIBRARY_PATH = File.expand_path("#{File.dirname(__FILE__)}/../../libexec")
    OVERMIND_PATH = "#{LIBRARY_PATH}/overmind"
    TMUX_FOLDER_PATH = "#{LIBRARY_PATH}/prebuilt-tmux/bin"
    TMUX_PATH = "#{TMUX_FOLDER_PATH}/tmux"

    def run(args = [])
      os_validate!
      validate_tmux_present!
      validate_overmind_present!

      # Ensures arguments are properly quoted if they contain spaces
      args = args.map { |x| x.include?(" ") ? "'#{x}'" : x }

      # Use prebuild tmux if found
      path_with_tmux = File.exist?(TMUX_PATH) ? "#{TMUX_FOLDER_PATH}:#{ENV["PATH"]}" : ENV["PATH"]

      # Spawns the Overmind process with modified PATH if necessary
      pid = spawn({"PATH" => path_with_tmux}, "#{OVERMIND_PATH} #{args.join(" ")}")

      Process.wait(pid)

      $?.exitstatus
    end

    private

    # Checks if the current OS is supported based on a regex match
    def os_supported?
      RUBY_PLATFORM.match?(SUPPORTED_OS_REGEX)
    end

    # Checks if tmux is installed either globally or as a prebuilt binary
    def tmux_installed?
      system("which tmux") || File.exist?(TMUX_PATH)
    end

    # Checks if the Overmind executable is present
    def overmind_installed?
      File.exist?(OVERMIND_PATH)
    end

    # Validates the operating system and aborts with an error message if unsupported
    def os_validate!
      return if os_supported?

      abort_with_message("Error: This gem supports Linux, *BSD, and macOS only.")
    end

    # Validates the presence of tmux and aborts with an error message if not found
    def validate_tmux_present!
      return if tmux_installed?

      abort_with_message(<<~MSG)
        Error: tmux not found. Please ensure tmux is installed and available in PATH.
        If tmux is not installed, you can usually install it using your package manager. For example:

          # For Ubuntu/Debian
          sudo apt-get install tmux
        #{"  "}
          # For macOS
          brew install tmux
        #{"  "}
          # For FreeBSD
          sudo pkg install tmux

        Installation commands might vary based on your operating system and its version.#{" "}
        Please consult your system's package management documentation for the most accurate instructions.
      MSG
    end

    # Validates the presence of Overmind and aborts with an error message if not found
    def validate_overmind_present!
      return if overmind_installed?

      abort_with_message("Error: Invalid platform. Overmind wasn't built for #{RUBY_PLATFORM}")
    end

    # Aborts execution with a given error message
    def abort_with_message(message)
      warn "\e[31m#{message}\e[0m"
      exit 1
    end
  end
end
