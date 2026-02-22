class Cbq < Formula
  desc "A clipboard manager that works like a stack or queue"
  homepage "https://github.com/vibe-coding/cbq"
  url "https://github.com/vibe-coding/cbq/archive/refs/tags/v0.1.0.tar.gz"
  sha256 "0000000000000000000000000000000000000000000000000000000000000000" # Placeholder
  license "MIT"

  depends_on "go" => :build

  def install
    system "go", "build", "-o", bin/"cbq", "main.go"
  end

  test do
    system "#{bin}/cbq", "--version"
  end
end
