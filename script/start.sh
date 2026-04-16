#!/bin/bash
#
# Script Name: start.sh
#
# Version:      1.1
# Author:       Naoki Hirata
# Date:         2026-04-05
# Usage:        start.sh [-test] [--help]
# Options:      -test      test mode execution with the latest source package
#               --help     show this help message
# Description:  This script builds Docker and Go sample environment by one-liner command.
# Version History:
#               1.0  (2026-03-13) Initial release
#               1.1  (2026-04-05) Fix SIGPIPE/pipefail issue when extracting archive directory name
# License:      MIT License

set -e
set -o pipefail

# Colors
readonly RED='\033[0;31m'
readonly GREEN='\033[0;32m'
readonly YELLOW='\033[1;33m'
readonly BLUE='\033[0;34m'
readonly BOLD='\033[1m'
readonly NC='\033[0m'

# Repository coordinates (change here when using another repository)
readonly REPO_USER="kdinstall"
readonly REPO_NAME="system-base5"

log_info() {
    echo -e "${GREEN}→${NC} $1"
}

log_warn() {
    echo -e "${YELLOW}!${NC} $1"
}

log_error() {
    echo -e "${RED}✗${NC} $1"
}

log_step() {
    echo -e "${BLUE}${BOLD}==>${NC}${BOLD} $1${NC}"
}

show_help() {
    cat <<EOF
oneliner-docker - 1行でDockerサーバ環境 + Goサンプル構築

Usage:
    curl -fsSL https://raw.githubusercontent.com/${REPO_USER}/${REPO_NAME}/master/script/start.sh | bash
  curl -fsSL ... | bash -s -- [-test] [--help]

Options:
  -test      Use latest master branch instead of latest release tag (for testing)
  --help     Show this help message

Target OS: Ubuntu 24

EOF
}

# Parse --help before other checks
for arg in "$@"; do
    case "$arg" in
        --help|-h)
            show_help
            exit 0
            ;;
    esac
done

GITHUB_USER="${REPO_USER}"
GITHUB_REPO="${REPO_NAME}"
readonly GITHUB_USER GITHUB_REPO
SCRIPT_URL="https://raw.githubusercontent.com/${GITHUB_USER}/${GITHUB_REPO}/master/script/start.sh"
readonly SCRIPT_URL

# Check os version
declare DIST_NAME=""
RELEASE_FILE=/etc/os-release

if [ -f "${RELEASE_FILE}" ] && grep -q '^NAME="Ubuntu' "${RELEASE_FILE}"; then
    DIST_NAME="Ubuntu"
fi

# Exit if unsupported os
if [ "${DIST_NAME}" == '' ]; then
    log_error "Your platform is not supported."
    uname -a
    exit 1
fi

# Define fixed parameters
readonly PLAYBOOKS=("docker" "app")
readonly WORK_DIR=/root/${GITHUB_REPO}_work
readonly INSTALL_PACKAGE_CMD="apt -y install"

# check root user
if [ "$(id -u)" -ne 0 ]; then
    log_error "This script must be run as root."
    echo
    echo "Please run with sudo:"
    echo "  curl -fsSL ${SCRIPT_URL} | sudo bash"
    exit 1
fi

log_step "${DIST_NAME} - START BUILDING ENVIRONMENT"

# Get test mode
if [ "$1" == '-test' ]; then
    readonly TEST_MODE=true
    log_info "Test mode: using latest master branch"
else
    readonly TEST_MODE=false
fi

# Install ansible command
if ! type -P ansible >/dev/null 2>&1; then
    log_step "Installing Ansible"
    ${INSTALL_PACKAGE_CMD} software-properties-common
    add-apt-repository --yes --update ppa:ansible/ansible
    ${INSTALL_PACKAGE_CMD} ansible-core
    log_info "Ansible installed"
else
    log_info "Ansible is already installed"
fi

# Download the latest repository archive
if ${TEST_MODE}; then
    url="https://github.com/${GITHUB_USER}/${GITHUB_REPO}/archive/master.tar.gz"
    version="new"
else
    set +e
    url=$(curl -sf "https://api.github.com/repos/${GITHUB_USER}/${GITHUB_REPO}/tags" 2>/dev/null | \
        grep '"tarball_url"' | head -n 1 | \
        sed -e 's/.*"tarball_url"[[:space:]]*:[[:space:]]*"\([^"]*\)".*/\1/')
    set -e
    if [ -z "${url}" ]; then
        log_error "Could not find release tag. Use -test to try latest master."
        exit 1
    fi
    version=$(basename "$url" | sed -e 's/v\([0-9\.]*\)/\1/')
    [ -z "${version}" ] && version="latest"
fi
filename=${GITHUB_REPO}_${version}.tar.gz
filepath=${WORK_DIR}/${filename}

# Set current directory
mkdir -p ${WORK_DIR}
cd ${WORK_DIR}
savefilelist=$(ls -1 2>/dev/null || true)

# Download archived repository
log_step "Downloading ${GITHUB_USER}/${GITHUB_REPO}"
if ! curl -fsSL -o "${filepath}" "${url}"; then
    log_error "Download failed: ${url}"
    exit 1
fi
if [ ! -s "${filepath}" ]; then
    log_error "Downloaded file is empty"
    exit 1
fi

# Remove old files
for file in $savefilelist; do
    [ -z "${file}" ] && continue
    if [ "${file}" != "${filename}" ]; then
        rm -rf "${file}"
    fi
done

# Get archive directory name
destdir=$(tar tzf "${filepath}" | head -n 1) || true
destdirname=$(basename "$destdir")
log_info "Extracting archive: ${filename}"

# Unarchive repository
tar xzf "${filename}"
find ./ -type f -name ".gitkeep" -delete
mv "${destdirname}" "${GITHUB_REPO}"
log_info "${filename} unarchived"

# launch ansible
for PLAYBOOK in "${PLAYBOOKS[@]}"; do
    log_step "Running Ansible playbook: ${PLAYBOOK}"
    
    playbook_dir="${WORK_DIR}/${GITHUB_REPO}/playbooks/${PLAYBOOK}"
    if [ ! -d "${playbook_dir}" ]; then
        log_error "Playbook directory not found: ${playbook_dir}"
        exit 1
    fi
    
    cd "${playbook_dir}" || exit 1
    
    if [ -f requirements.yml ]; then
        log_info "Installing Ansible galaxy requirements"
        ansible-galaxy install --role-file=requirements.yml || exit 1
    fi

    if [ ! -f main.yml ]; then
        log_error "Playbook main.yml not found in ${playbook_dir}"
        exit 1
    fi

    log_info "Executing ansible-playbook -i localhost, main.yml"
    if ! ansible-playbook -i localhost, main.yml; then
        log_error "Ansible playbook failed: ${PLAYBOOK}"
        exit 1
    fi
    
    log_info "${PLAYBOOK} playbook completed successfully"
done

log_step "Docker and Go sample environment setup complete"
