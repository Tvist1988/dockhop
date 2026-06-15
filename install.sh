#!/bin/sh
# Install the latest dockhop release.
# Usage: curl -fsSL https://raw.githubusercontent.com/Tvist1988/dockhop/master/install.sh | sh
set -e

REPO="Tvist1988/dockhop"
BIN="dockhop"

os=$(uname -s)
case "$os" in
  Linux) os="linux" ;;
  Darwin) os="darwin" ;;
  *) echo "Unsupported OS: $os" >&2; exit 1 ;;
esac

arch=$(uname -m)
case "$arch" in
  x86_64 | amd64) arch="amd64" ;;
  arm64 | aarch64) arch="arm64" ;;
  *) echo "Unsupported architecture: $arch" >&2; exit 1 ;;
esac

version=$(curl -fsSL "https://api.github.com/repos/${REPO}/releases/latest" \
  | grep '"tag_name":' | head -n1 | cut -d'"' -f4)
if [ -z "$version" ]; then
  echo "Could not determine the latest release of ${REPO}" >&2
  exit 1
fi

num=${version#v}
url="https://github.com/${REPO}/releases/download/${version}/${BIN}_${num}_${os}_${arch}.tar.gz"

tmp=$(mktemp -d)
trap 'rm -rf "$tmp"' EXIT

echo "Downloading ${BIN} ${version} (${os}/${arch})..."
curl -fsSL "$url" -o "$tmp/${BIN}.tar.gz"
tar -xzf "$tmp/${BIN}.tar.gz" -C "$tmp"

dir="/usr/local/bin"
if [ ! -w "$dir" ]; then
  dir="$HOME/.local/bin"
  mkdir -p "$dir"
fi

install -m 0755 "$tmp/${BIN}" "$dir/${BIN}"
echo "Installed ${BIN} to ${dir}/${BIN}"
case ":$PATH:" in
  *":$dir:"*) ;;
  *) echo "Note: ${dir} is not in your PATH — add it to use '${BIN}' directly." ;;
esac
