#!/bin/sh
#
# This script is for installing the kubectl-image plugin.
# It automatically detects the OS and architecture, then downloads the
# latest release from GitHub and installs it.
#
# Usage:
#   curl -fsSL https://raw.githubusercontent.com/reedchan7/kubectl-image/main/install.sh | sh
#
# To install a specific version, set the VERSION environment variable:
#   VERSION=v0.2.0 curl -fsSL ... | sh

set -e

# Define the project and repository
PROJECT_NAME="kubectl-image"
REPO="reedchan7/kubectl-image"

# Determine the version to install
if [ -z "$VERSION" ]; then
  # Get the latest release version from GitHub API
  VERSION=$(curl -s "https://api.github.com/repos/$REPO/releases/latest" | grep '"tag_name":' | sed -E 's/.*"([^"]+)".*/\1/')
  if [ -z "$VERSION" ]; then
    echo "Error: Could not determine the latest version. Please set the VERSION environment variable."
    exit 1
  fi
fi

echo "Installing $PROJECT_NAME version $VERSION..."

# Detect OS and architecture
OS="$(uname -s | tr '[:upper:]' '[:lower:]')"
ARCH="$(uname -m)"
case "$ARCH" in
  x86_64 | amd64)
    ARCH="amd64"
    ;;
  armv8* | aarch64* | arm64)
    ARCH="arm64"
    ;;
  *)
    echo "Error: Unsupported architecture: $ARCH"
    exit 1
    ;;
esac

# Construct the download URL
DOWNLOAD_URL="https://github.com/$REPO/releases/download/$VERSION/${PROJECT_NAME}_${OS}_${ARCH}.tar.gz"

# Create a temporary directory for the download
TMP_DIR=$(mktemp -d)
cd "$TMP_DIR"

# Download and extract the archive
echo "Downloading from $DOWNLOAD_URL"
if ! curl -fsSLO "$DOWNLOAD_URL"; then
  echo "Error: Failed to download $DOWNLOAD_URL"
  exit 1
fi

echo "Extracting archive..."
tar -zxf "${PROJECT_NAME}_${OS}_${ARCH}.tar.gz"

# Verify the binary exists
if [ ! -f "./$PROJECT_NAME" ]; then
  echo "Error: Binary $PROJECT_NAME not found in archive"
  ls -la
  exit 1
fi

# Install the binary
echo "Installing to /usr/local/bin (requires sudo)"
sudo mv "./$PROJECT_NAME" /usr/local/bin/

# Clean up
cd ..
rm -rf "$TMP_DIR"

echo "$PROJECT_NAME has been installed successfully. You can now use it as 'kubectl image'."
