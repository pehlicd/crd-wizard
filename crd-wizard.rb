class CrdWizard < Formula
  desc "CR(D) Wizard is a tool to explore Kubernetes CRDs via a TUI or web interface."
  homepage "https://github.com/pehlicd/crd-wizard"
  url "https://github.com/pehlicd/crd-wizard.git",
      tag:      "v0.1.9",
      revision: "70f1a97810ef984ce632773dcb73ea2575b19720"
  license "GPL-3.0"
  head "https://github.com/pehlicd/crd-wizard.git", branch: "main"

  depends_on "go" => :build

  def install
    project = "github.com/pehlicd/crd-wizard"
    ldflags = %W[
      -s -w
      -X #{project}/cmd.versionString=#{version}
      -X #{project}/cmd.buildCommit=#{Utils.git_head}
      -X #{project}/cmd.buildDate=#{time.iso8601}
    ]
    system "go", "build", *std_go_args(ldflags: ldflags)
  end

  test do
    assert_match version.to_s, "#{bin}/crd-wizard version"
  end
end
