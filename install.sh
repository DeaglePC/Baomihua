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

SUDO_CMD=""
# Check if we need sudo to write to INSTALL_DIR (or create it)
if [ ! -d "$INSTALL_DIR" ] || [ ! -w "$INSTALL_DIR" ]; then
  SUDO_CMD="sudo"
  echo "Elevating privileges to install to $INSTALL_DIR (you may be prompted for password)..."
fi

# Ensure directory exists
if [ ! -d "$INSTALL_DIR" ]; then
  $SUDO_CMD mkdir -p "$INSTALL_DIR"
fi

echo "Installing bmh to $INSTALL_DIR..."
chmod +x bmh
$SUDO_CMD mv bmh "$INSTALL_DIR/bmh"

echo "Installation complete."
echo "Running initialization..."
"$INSTALL_DIR/bmh" install
