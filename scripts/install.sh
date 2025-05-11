#!/bin/bash
#
# install.sh - Installer for the 'nicksh' CLI tool
#
# Usage:
#   curl -sSL https://raw.githubusercontent.com/AntonioJCosta/nicksh/main/scripts/install.sh | bash
#   wget -qO- https://raw.githubusercontent.com/AntonioJCosta/nicksh/main/scripts/install.sh | bash
#
# Or, to install a specific version:
#   curl -sSL https://raw.githubusercontent.com/AntonioJCosta/nicksh/main/scripts/install.sh | bash -s -- -v v1.0.0
#
# The script will attempt to install to /usr/local/bin,
# if it fails due to permissions, it will try $HOME/.local/bin.
#
# Prerequisites: curl, tar, gzip

# Strict mode
set -e

# --- Configuration ---
GITHUB_USER="AntonioJCosta"
GITHUB_REPO="nicksh"
INSTALL_NAME="nicksh"

# --- Colors ---
Color_Off='\033[0m'
# Bold
BRed='\033[1;31m'; BGreen='\033[1;32m'; BYellow='\033[1;33m'; BPurple='\033[1;35m'; BCyan='\033[1;36m';

# --- Helper Functions ---
echo_info() {
  echo -e "${BCyan}[INFO]${Color_Off} $1"
}

echo_warn() {
  echo -e "${BYellow}[WARN]${Color_Off} $1"
}

echo_error() {
  echo -e "${BRed}[ERROR]${Color_Off} $1" >&2
  exit 1
}

echo_success() {
  echo -e "${BGreen}[SUCCESS]${Color_Off} $1"
}

# --- Dependency Checks ---
check_command() {
  if ! command -v "$1" &>/dev/null; then
    echo_error "Required command '$1' not found. Please install it and try again."
  fi
}

# --- Determine OS and Architecture ---
get_os_arch() {
  OS_TYPE=$(uname -s | tr '[:upper:]' '[:lower:]')
  CPU_ARCH=$(uname -m)

  case "$OS_TYPE" in
    linux|darwin)
      # OS_TYPE is already set correctly
      ;;
    *)
      echo_error "Unsupported operating system: ${OS_TYPE}. Only Linux and macOS are supported."
      ;;
  esac

  case "$CPU_ARCH" in
    x86_64|amd64)
      CPU_ARCH="amd64"
      ;;
    arm64|aarch64)
      CPU_ARCH="arm64"
      ;;
    *)
      echo_error "Unsupported architecture: ${CPU_ARCH}. Only amd64 (x86_64) and arm64 (aarch64) are supported."
      ;;
  esac
  echo_info "Detected OS: ${OS_TYPE}, Arch: ${CPU_ARCH}"
}

# --- Get Release Version ---
get_release_version() {
  if [ -n "$TARGET_VERSION" ]; then
    LATEST_RELEASE_TAG="$TARGET_VERSION"
    echo_info "Targeting specified version: ${BYellow}${LATEST_RELEASE_TAG}${Color_Off}"
  else
    echo_info "Fetching the latest release tag from GitHub..."
    # Get the final URL after redirects from /releases/latest
    # This URL will be in the format: https://github.com/USER/REPO/releases/tag/vX.Y.Z
    local latest_release_url
    latest_release_url=$(curl -Ls -o /dev/null -w "%{url_effective}" "https://github.com/${GITHUB_USER}/${GITHUB_REPO}/releases/latest")

    if [ -z "$latest_release_url" ]; then
      echo_error "Could not determine the latest release URL. Please check the repository and network."
    fi

    LATEST_RELEASE_TAG=$(basename "$latest_release_url")

    if [ -z "$LATEST_RELEASE_TAG" ] || [[ "$LATEST_RELEASE_TAG" == "latest" ]] || [[ "$LATEST_RELEASE_TAG" == "releases" ]]; then
      echo_error "Could not extract a valid release tag from URL: ${latest_release_url}. Does the repository have releases?"
    fi
    echo_info "Latest release tag: ${BGreen}${LATEST_RELEASE_TAG}${Color_Off}"
  fi
}

# --- Parse Arguments (for specific version) ---
parse_arguments() {
  TARGET_VERSION=""
  while [ "$#" -gt 0 ]; do
    case "$1" in
      -v|--version)
        if [ -z "$2" ]; then
          echo_error "Option $1 requires an argument."
        fi
        TARGET_VERSION="$2"
        shift 2
        ;;
      *)
        echo_error "Unknown option: $1"
        ;;
    esac
  done
}

# --- Main Installation Logic ---
main() {
  parse_arguments "$@"

  check_command "curl"
  check_command "tar"
  # gzip is usually implicitly handled by tar for .tar.gz
  # check_command "gzip" 

  get_os_arch
  get_release_version

  local binary_filename_base="${INSTALL_NAME}-${OS_TYPE}-${CPU_ARCH}"
  local archive_filename="${binary_filename_base}-${LATEST_RELEASE_TAG}.tar.gz"
  local download_url="https://github.com/${GITHUB_USER}/${GITHUB_REPO}/releases/download/${LATEST_RELEASE_TAG}/${archive_filename}"

  local tmp_dir
  tmp_dir=$(mktemp -d -t "${INSTALL_NAME}_install.XXXXXX")
  trap 'echo_info "Cleaning up temporary directory..."; rm -rf "$tmp_dir" && echo_success "Temporary directory cleaned successfully!" || echo_warn "Failed to clean up temporary directory: $tmp_dir"' EXIT ERR INT TERM

  echo_info "Downloading ${BYellow}${archive_filename}${Color_Off} from ${download_url}..."
  if ! curl --progress-bar -L "$download_url" -o "$tmp_dir/$archive_filename"; then
    echo_error "Download failed. Please check the URL (${download_url}) and your network connection."
  fi

  echo_info "Extracting ${BYellow}${archive_filename}${Color_Off}..."
  if ! tar -xzf "$tmp_dir/$archive_filename" -C "$tmp_dir"; then
    echo_error "Extraction failed. The archive might be corrupted or in an unexpected format."
  fi
  
  local extracted_binary_name="${binary_filename_base}" 
  if [ ! -f "$tmp_dir/$extracted_binary_name" ]; then
    echo_error "Extracted binary '${extracted_binary_name}' not found in the archive."
  fi

  echo_info "Attempting to install ${INSTALL_NAME}..."
  
  local final_install_dir=""
  local install_dir_system="/usr/local/bin"
  local install_dir_user="$HOME/.local/bin"

  # Try system-wide installation first
  if [ -w "$install_dir_system" ] || { sudo -n true 2>/dev/null && [ -d "$install_dir_system" ]; } ; then
    echo_info "Attempting to install to ${install_dir_system} (may require sudo)..."
    if [ -f "$install_dir_system/$INSTALL_NAME" ]; then
      echo_info "An existing version of ${INSTALL_NAME} was found at $install_dir_system/$INSTALL_NAME and will be replaced."
    fi
    if sudo mv "$tmp_dir/$extracted_binary_name" "$install_dir_system/$INSTALL_NAME"; then
      sudo chmod +x "$install_dir_system/$INSTALL_NAME"
      echo_success "${INSTALL_NAME} installed successfully to $install_dir_system/$INSTALL_NAME"
      final_install_dir="$install_dir_system"
    else
      echo_warn "Failed to install to ${install_dir_system} with sudo. Trying user installation."
    fi
  else
    echo_warn "${install_dir_system} is not writable or sudo is not configured for non-interactive use. Trying user installation."
  fi

  # If system-wide installation failed or wasn't attempted, try user-specific installation
  if [ -z "$final_install_dir" ]; then
    echo_info "Attempting to install to ${install_dir_user}..."
    mkdir -p "$install_dir_user"
    if [ -f "$install_dir_user/$INSTALL_NAME" ]; then
      echo_info "An existing version of ${INSTALL_NAME} was found at $install_dir_user/$INSTALL_NAME and will be replaced."
    fi
    if mv "$tmp_dir/$extracted_binary_name" "$install_dir_user/$INSTALL_NAME"; then
      chmod +x "$install_dir_user/$INSTALL_NAME"
      echo_success "${INSTALL_NAME} installed successfully to $install_dir_user/$INSTALL_NAME"
      final_install_dir="$install_dir_user"
    else
      echo_error "Failed to install to ${install_dir_user}. Please check permissions or try manual installation."
    fi
  fi


  if [ -z "$final_install_dir" ]; then
    echo_error "Installation could not be completed in any standard location."
    # This case should ideally not be reached if the logic above is correct
  fi

  echo_info ""
  echo_success "Installation complete!"
  echo_info "You can now try running: ${BYellow}${INSTALL_NAME} --version${Color_Off}"
  echo_info ""
  echo_info "${BPurple}Optional:${Color_Off} For enhanced interactive features (like alias selection and history search),"
  echo_info "install 'fzf'. Common ways to install fzf:"
  echo_info "  - macOS (Homebrew): ${BYellow}brew install fzf${Color_Off}"
  echo_info "  - Debian/Ubuntu:    ${BYellow}sudo apt install fzf${Color_Off}"
  echo_info "  - Fedora:           ${BYellow}sudo dnf install fzf${Color_Off}"
  echo_info "  - Arch Linux:       ${BYellow}sudo pacman -S fzf${Color_Off}"
  echo_info "  - From source:      ${BYellow}git clone --depth 1 https://github.com/junegunn/fzf.git ~/.fzf && ~/.fzf/install${Color_Off}"
  echo_info ""
  
  if ! command -v "$INSTALL_NAME" &>/dev/null || [[ ":$PATH:" != *":${final_install_dir}:"* ]]; then
    echo_warn "The installation directory (${BYellow}${final_install_dir}${Color_Off}) might not be in your PATH or your shell hasn't picked up the change yet."
    echo_warn "You may need to add it to your shell configuration file (e.g., ~/.bashrc, ~/.zshrc):"
    echo_warn "  ${BYellow}echo 'export PATH=\"${final_install_dir}:\$PATH\"' >> ~/.your_shell_rc_file${Color_Off}"
    echo_warn "Then, source your shell config (e.g., ${BYellow}source ~/.your_shell_rc_file${Color_Off}) or open a new terminal."
  fi
}

# --- Run Script ---
main "$@"