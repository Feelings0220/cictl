#!/usr/bin/env bash
# Install the cictl binary for this machine from GitHub Releases.
#
# The skill deliberately does NOT bundle the binary (it is platform-specific and
# would bloat the repo). This script downloads the archive that matches your
# OS/arch, verifies its SHA-256 against the release checksums.txt, and installs
# cictl to BIN_DIR (default: ~/.local/bin).
#
# Usage:
#   bash install-cictl.sh            # latest release
#   bash install-cictl.sh v0.1.1     # a specific tag
#   BIN_DIR=/usr/local/bin bash install-cictl.sh
set -euo pipefail

REPO="Feelings0220/cictl"
BIN_DIR="${BIN_DIR:-$HOME/.local/bin}"
TAG="${1:-}"

need() { command -v "$1" >/dev/null 2>&1 || { echo "error: '$1' is required but not found" >&2; exit 1; }; }
need curl
need tar

# --- resolve tag (default: latest) ---
if [ -z "$TAG" ]; then
  TAG="$(curl -fsSL "https://api.github.com/repos/$REPO/releases/latest" \
    | grep -m1 '"tag_name"' | sed -E 's/.*"tag_name": *"([^"]+)".*/\1/')"
fi
[ -n "$TAG" ] || { echo "error: could not resolve a release tag" >&2; exit 1; }
VERSION="${TAG#v}"   # goreleaser archive names drop the leading 'v'

# --- detect OS / arch ---
case "$(uname -s)" in
  Linux)  OS=linux ;;
  Darwin) OS=darwin ;;
  MINGW*|MSYS*|CYGWIN*) OS=windows ;;
  *) echo "error: unsupported OS '$(uname -s)'" >&2; exit 1 ;;
esac
case "$(uname -m)" in
  x86_64|amd64) ARCH=amd64 ;;
  aarch64|arm64) ARCH=arm64 ;;
  *) echo "error: unsupported arch '$(uname -m)'" >&2; exit 1 ;;
esac

if [ "$OS" = windows ]; then EXT=zip; BIN=cictl.exe; else EXT=tar.gz; BIN=cictl; fi
ARCHIVE="cictl_${VERSION}_${OS}_${ARCH}.${EXT}"
BASE="https://github.com/$REPO/releases/download/$TAG"

echo "Installing cictl $TAG for ${OS}/${ARCH} -> $BIN_DIR"

TMP="$(mktemp -d)"
trap 'rm -rf "$TMP"' EXIT

curl -fsSL "$BASE/$ARCHIVE"       -o "$TMP/$ARCHIVE"
curl -fsSL "$BASE/checksums.txt"  -o "$TMP/checksums.txt"

# --- verify checksum ---
EXPECTED="$(grep " $ARCHIVE\$" "$TMP/checksums.txt" | awk '{print $1}')"
[ -n "$EXPECTED" ] || { echo "error: $ARCHIVE not listed in checksums.txt" >&2; exit 1; }
if command -v sha256sum >/dev/null 2>&1; then
  ACTUAL="$(sha256sum "$TMP/$ARCHIVE" | awk '{print $1}')"
else
  ACTUAL="$(shasum -a 256 "$TMP/$ARCHIVE" | awk '{print $1}')"
fi
[ "$EXPECTED" = "$ACTUAL" ] || { echo "error: checksum mismatch for $ARCHIVE" >&2; exit 1; }

# --- extract + install ---
if [ "$EXT" = zip ]; then
  need unzip
  unzip -q "$TMP/$ARCHIVE" -d "$TMP"
else
  tar -xzf "$TMP/$ARCHIVE" -C "$TMP"
fi
mkdir -p "$BIN_DIR"
install -m 0755 "$TMP/$BIN" "$BIN_DIR/$BIN" 2>/dev/null || { cp "$TMP/$BIN" "$BIN_DIR/$BIN"; chmod 0755 "$BIN_DIR/$BIN"; }

echo "Installed: $BIN_DIR/$BIN"
case ":$PATH:" in
  *":$BIN_DIR:"*) : ;;
  *) echo "note: $BIN_DIR is not on PATH — add it, e.g. export PATH=\"$BIN_DIR:\$PATH\"" ;;
esac
"$BIN_DIR/$BIN" --version || true
