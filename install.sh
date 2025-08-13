#!/usr/bin/env bash
#
# The crd-wizard universal installer for Linux and macOS.

set -e
set -o pipefail

GITHUB_USER="pehlicd"
GITHUB_REPO="crd-wizard"
BINARY_NAME="crd-wizard"
INSTALL_DIR_DEFAULT="/usr/local/bin"

setup_colors() {
  if [ -t 1 ]; then
    BLUE=$(printf '\033[34m')
    GREEN=$(printf '\033[32m')
    RED=$(printf '\033[31m')
    YELLOW=$(printf '\033[33m')
    BOLD=$(printf '\033[1m')
    RESET=$(printf '\033[0m')
  else
    BLUE=""
    GREEN=""
    RED=""
    YELLOW=""
    BOLD=""
    RESET=""
  fi
}

info() {
    printf "${YELLOW}[INFO]${RESET} %s\\n" "$@"
}

fatal() {
    printf "${RED}[ERROR]${RESET} %s\\n" "$@" >&2
    exit 1
}

verify_command() {
    command -v "$1" >/dev/null 2>&1
}

setup_platform() {
    OS=$(uname -s | tr '[:upper:]' '[:lower:]')
    ARCH=$(uname -m)

    case "$OS" in
        linux) ;;
        darwin) ;;
        *) fatal "Unsupported OS: $OS. This script supports Linux and macOS." ;;
    esac

    case "$ARCH" in
        x86_64) ARCH="amd64" ;;
        aarch64 | arm64) ARCH="arm64" ;;
        *) fatal "Unsupported architecture: $ARCH" ;;
    esac
    info "Detected Platform: ${BOLD}$OS/$ARCH${RESET}"
}

setup_tmp_dir() {
    TMP_DIR=$(mktemp -d -t crd-wizard-install.XXXXXXXXXX)
    cleanup() {
        local code=$?
        set +e
        info "Cleaning up temporary directory..."
        rm -rf "$TMP_DIR"
        exit $code
    }
    trap cleanup INT TERM EXIT
}

setup_downloader() {
    if verify_command "curl"; then
        DOWNLOADER="curl"
    elif verify_command "wget"; then
        DOWNLOADER="wget"
    else
        fatal "Cannot find 'curl' or 'wget'. Please install one of them."
    fi
    info "Using ${BOLD}$DOWNLOADER${RESET} for downloads"
}

download_file() {
    local url="$1"
    local dest="$2"
    info "Downloading ${BLUE}$url${RESET}"

    case $DOWNLOADER in
        curl) curl -sfL -o "$dest" "$url" ;;
        wget) wget -qO "$dest" "$url" ;;
    esac

    if [ $? -ne 0 ]; then
        fatal "Download failed for $url. Please check the URL and your network."
    fi
}

get_release_info() {
    local release_api_url
    if [ -n "$CRD_WIZARD_VERSION" ]; then
        release_api_url="https://api.github.com/repos/$GITHUB_USER/$GITHUB_REPO/releases/tags/v$CRD_WIZARD_VERSION"
    else
        release_api_url="https://api.github.com/repos/$GITHUB_USER/$GITHUB_REPO/releases/latest"
    fi

    info "Fetching release information from GitHub API..."
    RELEASE_TAG=$(curl -s "$release_api_url" | grep '"tag_name":' | sed -E 's/.*"([^"]+)".*/\1/')

    if [ -z "$RELEASE_TAG" ]; then
        fatal "Could not determine release version. Please check the repository and CRD_WIZARD_VERSION."
    fi

    info "Installing version: ${BOLD}$RELEASE_TAG${RESET}"

    local version_without_v=${RELEASE_TAG#v}
    local download_base_url="https://github.com/$GITHUB_USER/$GITHUB_REPO/releases/download/$RELEASE_TAG"

    ASSET_NAME="${GITHUB_REPO}_${RELEASE_TAG}_${OS}_${ARCH}.tar.gz"
    CHECKSUMS_NAME="${GITHUB_REPO}_${version_without_v}_checksums.txt"

    ASSET_URL="$download_base_url/$ASSET_NAME"
    CHECKSUMS_URL="$download_base_url/$CHECKSUMS_NAME"
}

verify_checksum() {
    local tarball_path="$1"
    local checksums_path="$2"
    local asset_name="$3"

    info "Verifying checksum for ${BOLD}$asset_name${RESET}..."
    if ! verify_command "sha256sum" && ! verify_command "shasum"; then
        fatal "Cannot find 'sha256sum' or 'shasum' to verify download."
    fi

    local expected_hash
    expected_hash=$(grep "$asset_name" "$checksums_path" | awk '{print $1}')

    if [ -z "$expected_hash" ]; then
        fatal "Could not find checksum for '$asset_name' in $checksums_path."
    fi

    local actual_hash
    if verify_command "sha256sum"; then
        actual_hash=$(sha256sum "$tarball_path" | awk '{print $1}')
    else
        actual_hash=$(shasum -a 256 "$tarball_path" | awk '{print $1}')
    fi

    if [ "$expected_hash" != "$actual_hash" ]; then
        fatal "Checksum mismatch! Expected '$expected_hash' but got '$actual_hash'."
    fi
    info "${GREEN}Checksum verified successfully.${RESET}"
}

main() {
    setup_colors
    INSTALL_DIR="${1:-$INSTALL_DIR_DEFAULT}"

    setup_platform
    setup_downloader
    setup_tmp_dir
    get_release_info

    download_file "$CHECKSUMS_URL" "$TMP_DIR/$CHECKSUMS_NAME"
    download_file "$ASSET_URL" "$TMP_DIR/$ASSET_NAME"

    verify_checksum "$TMP_DIR/$ASSET_NAME" "$TMP_DIR/$CHECKSUMS_NAME" "$ASSET_NAME"

    info "Extracting binary..."
    tar -xzf "$TMP_DIR/$ASSET_NAME" -C "$TMP_DIR"

    info "Installing ${BOLD}$BINARY_NAME${RESET} to ${BOLD}$INSTALL_DIR${RESET}..."

    if [ ! -w "$INSTALL_DIR" ]; then
        if sudo -n true 2>/dev/null; then
            sudo mv "$TMP_DIR/$BINARY_NAME" "$INSTALL_DIR/$BINARY_NAME"
        else
            info "Write permissions not found."
            printf "Please enter your password to complete the installation with sudo.\\n"
            sudo mv "$TMP_DIR/$BINARY_NAME" "$INSTALL_DIR/$BINARY_NAME"
        fi
    else
        mv "$TMP_DIR/$BINARY_NAME" "$INSTALL_DIR/$BINARY_NAME"
    fi

    printf "\\n"
    printf "${GREEN}✅ %s was installed successfully!${RESET}\\n" "${BOLD}${BINARY_NAME}${RESET}"
    printf "   Run '%s' to get started.\\n" "${BOLD}${BINARY_NAME} --help${RESET}"
}

main "$@"
