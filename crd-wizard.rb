class CrdWizard < Formula
  desc "CR(D) Wizard is a tool to explore Kubernetes CRDs via a TUI or web interface."
  homepage "https://github.com/pehlicd/crd-wizard"
  url "https://github.com/pehlicd/crd-wizard.git",
      tag:      "v0.1.5",
      revision: "3e47d680efd49b02873a13b79288cb087d49ff85"
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
