#!/bin/bash
#
# uninstall.sh - Uninstaller for the 'nicksh' CLI tool
#
# Usage:
#   curl -sSL https://raw.githubusercontent.com/AntonioJCosta/nicksh/main/scripts/uninstall.sh | bash
#   wget -qO- https://raw.githubusercontent.com/AntonioJCosta/nicksh/main/scripts/uninstall.sh | bash

# Strict mode
set -euo pipefail

# --- Configuration ---
INSTALL_NAME="nicksh"
CONFIG_DIR="$HOME/.nicksh"

# --- Colors ---
Color_Off='\033[0m'
BRed='\033[1;31m'; BGreen='\033[1;32m'; BYellow='\033[1;33m'; BCyan='\033[1;36m';

# --- Helper Functions ---
echo_info() {
  echo -e "${BCyan}[INFO]${Color_Off} $1"
}

echo_warn() {
  echo -e "${BYellow}[WARN]${Color_Off} $1"
}

echo_error() {
  echo -e "${BRed}[ERROR]${Color_Off} $1" >&2
}

echo_success() {
  echo -e "${BGreen}[SUCCESS]${Color_Off} $1"
}

# --- Main Uninstall Logic ---
main() {
  echo_info "Attempting to uninstall ${INSTALL_NAME}..."
  echo_info "This script will attempt to remove the ${INSTALL_NAME} binary and optionally its configuration directory."
  echo_info ""

  local install_dir_system="/usr/local/bin"
  local install_dir_user="$HOME/.local/bin"
  local uninstalled_binary=false
  local removed_config=false

  local system_path="${install_dir_system}/${INSTALL_NAME}"
  local user_path="${install_dir_user}/${INSTALL_NAME}"

  if [ -f "$system_path" ]; then
    echo_info "${INSTALL_NAME} binary found at ${system_path}."
    echo_info "Attempting to remove (may require sudo)..."
    if sudo rm -f "$system_path"; then
      echo_success "${INSTALL_NAME} binary successfully uninstalled from ${system_path}."
      uninstalled_binary=true
    else
      echo_warn "Failed to remove ${INSTALL_NAME} binary from ${system_path}."
      echo_warn "You may need to remove it manually with sudo: sudo rm -f ${system_path}"
    fi
  elif [ -f "$user_path" ]; then
    echo_info "${INSTALL_NAME} binary found at ${user_path}."
    echo_info "Attempting to remove..."
    if rm -f "$user_path"; then
      echo_success "${INSTALL_NAME} binary successfully uninstalled from ${user_path}."
      uninstalled_binary=true
    else
      echo_error "Failed to remove ${INSTALL_NAME} binary from ${user_path}. Please check permissions."
    fi
  else
    echo_warn "${INSTALL_NAME} binary not found in standard installation locations (${system_path} or ${user_path})."
  fi
  echo_info ""

  if [ -d "$CONFIG_DIR" ]; then
    echo_info "Configuration directory found at ${CONFIG_DIR} (contains your aliases)."
    read -r -p "Do you want to remove this configuration directory? [y/N]: " confirmation
    if [[ "$confirmation" =~ ^[Yy]$ ]]; then
      echo_info "Attempting to remove configuration directory ${CONFIG_DIR}..."
      if rm -rf "$CONFIG_DIR"; then
        echo_success "Configuration directory ${CONFIG_DIR} successfully removed."
        removed_config=true
      else
        echo_warn "Failed to remove configuration directory ${CONFIG_DIR}. Please check permissions and remove it manually if needed: rm -rf ${CONFIG_DIR}"
      fi
    else
      echo_info "Skipped removal of configuration directory ${CONFIG_DIR}."
    fi
  else
    echo_info "Configuration directory ${CONFIG_DIR} not found. No action needed for configuration."
  fi

  echo_info ""
  echo_info "--- Uninstall Summary ---"
  local actions_taken=false

  if [ "$uninstalled_binary" = true ]; then
    echo_success "Binary: REMOVED"
    actions_taken=true
  else
    echo_info "Binary: NOT REMOVED (or was not found, see details above)"
  fi

  if [ "$removed_config" = true ]; then
    echo_success "Configuration (~/.nicksh): REMOVED"
    actions_taken=true
  else
    if [ -d "$CONFIG_DIR" ]; then
        echo_info "Configuration (~/.nicksh): NOT REMOVED (still exists, see details above)"
    else
        echo_info "Configuration (~/.nicksh): NOT PRESENT (or was not found initially, see details above)"
    fi
  fi
  
  echo_info "-------------------------"
  if [ "$actions_taken" = true ]; then
    echo_info "Uninstallation process has completed."
  elif ! [ -f "$system_path" ] && ! [ -f "$user_path" ] && ! [ -d "$CONFIG_DIR" ]; then
    echo_info "Nothing to uninstall. ${INSTALL_NAME} components were not found."
  else
    echo_warn "Review the messages above. Some components may not have been removed or found."
  fi
}

# --- Run Script ---
main "$@"