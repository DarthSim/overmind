# frozen_string_literal: true

require_relative "lib/overmind/version"

Gem::Specification.new do |spec|
  spec.name = "overmind"
  spec.version = Overmind::VERSION
  spec.authors = ["prog-supdex"]
  spec.email = ["symeton@gmail.com"]

  spec.summary = "Overmind is a process manager for Procfile-based applications and tmux."
  spec.description = "Overmind is a process manager for Procfile-based applications and tmux."
  spec.homepage = "https://github.com/DarthSim/overmind"
  spec.license = "MIT"
  spec.required_ruby_version = ">= 2.3"

  spec.metadata["homepage_uri"] = spec.homepage
  spec.metadata["source_code_uri"] = "https://github.com/DarthSim/overmind"

  # Specify which files should be added to the gem when it is released.
  spec.files = Dir["bin/*", "lib/**/*.rb", "libexec/**/*", "overmind.gemspec", "LICENSE.txt"]
  spec.bindir = "bin"
  spec.executables = ["overmind"]
  spec.require_paths = %w[lib libexec]

  spec.add_development_dependency "bundler"
  spec.add_development_dependency "rake", "~> 13.0"
  spec.add_development_dependency "rspec", "~> 3.5"
end
