#!/usr/bin/env bash
set -e

echo "Starting BaoMiHua (bmh) installation..."

# Detect OS and architecture
OS="$(uname -s | tr '[:upper:]' '[:lower:]')"
ARCH="$(uname -m)"

case "$ARCH" in
  x86_64) ARCH="amd64" ;;
  aarch64|arm64) ARCH="arm64" ;;
  *)
    echo "Unsupported architecture: $ARCH"
    exit 1
    ;;
esac

if [ "$OS" != "linux" ] && [ "$OS" != "darwin" ]; then
  echo "Unsupported operating system: $OS"
  exit 1
fi

REPO="DeaglePC/Baomihua"
FILE="bmh-${OS}-${ARCH}"

echo "Detected system: $OS, Architecture: $ARCH"
echo "Fetching latest release from $REPO..."

# Fetch latest release URL
DOWNLOAD_URL=$(curl -s "https://api.github.com/repos/$REPO/releases/latest" | grep -o "https://.*releases/download/.*/$FILE")

if [ -z "$DOWNLOAD_URL" ]; then
  echo "Failed to retrieve the download URL. The release might not exist for $OS/$ARCH."
  exit 1
fi

echo "Downloading $FILE..."
curl -L -o bmh "$DOWNLOAD_URL"

# Install binary
INSTALL_DIR="/usr/local/bin"
if [ ! -w "$INSTALL_DIR" ]; then
  echo "No write permission for $INSTALL_DIR. Attempting to install to ~/.local/bin"
  INSTALL_DIR="$HOME/.local/bin"
  mkdir -p "$INSTALL_DIR"
  if [[ ":$PATH:" != *":$INSTALL_DIR:"* ]]; then
    echo "WARNING: $INSTALL_DIR is not in your PATH. Please add it manually."
    echo 'Export PATH=$HOME/.local/bin:$PATH in your shell configuration.'
  fi
fi

echo "Installing bmh to $INSTALL_DIR..."
chmod +x bmh
mv bmh "$INSTALL_DIR/bmh"

echo "Installation complete."
echo "Running initialization..."
"$INSTALL_DIR/bmh" install

