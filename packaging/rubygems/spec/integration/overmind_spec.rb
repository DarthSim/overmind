# frozen_string_literal: true

require "spec_helper"
require "fileutils"

LOG_FILE_PATH = "tmp/overmind.log"
PROFILE_PATH = "spec/dummy/Procfile"
OVERMIND_SOCK = "./.overmind.sock"

RSpec.describe "Overmind", type: :feature do
  before(:all) do
    FileUtils.mkdir_p("tmp")

    FileUtils.touch(LOG_FILE_PATH)

    system("overmind start -f #{PROFILE_PATH} > #{LOG_FILE_PATH} 2>&1")
  end

  after(:all) do
    File.delete(OVERMIND_SOCK) if File.exist?(OVERMIND_SOCK)
    File.delete(LOG_FILE_PATH) if File.exist?(LOG_FILE_PATH)
  end

  it "logs expected output from the process" do
    log_content = File.read(LOG_FILE_PATH)
    expect(log_content).to include("Server is running")
  end
end
