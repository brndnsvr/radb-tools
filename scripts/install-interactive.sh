#!/bin/bash
# Interactive installation script for RADb Client
# This script builds, installs, and configures radb-client

set -e

# Color output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Logging functions
info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Detect OS and architecture
detect_platform() {
    OS=$(uname -s | tr '[:upper:]' '[:lower:]')
    ARCH=$(uname -m)

    case "$ARCH" in
        x86_64) ARCH="amd64" ;;
        aarch64|arm64) ARCH="arm64" ;;
        *)
            error "Unsupported architecture: $ARCH"
            exit 1
            ;;
    esac

    info "Detected platform: ${OS}/${ARCH}"
}

# Check prerequisites
check_prerequisites() {
    info "Checking prerequisites..."

    # Check for Go
    if ! command -v go &> /dev/null; then
        error "Go is not installed. Please install Go 1.23 or higher."
        error "Visit: https://golang.org/doc/install"
        exit 1
    fi

    GO_VERSION=$(go version | awk '{print $3}' | sed 's/go//')
    info "Found Go version: $GO_VERSION"

    # Check for git (for version info)
    if ! command -v git &> /dev/null; then
        warning "Git is not installed. Version information will be limited."
    fi

    success "Prerequisites check passed"
}

# Build the binary
build_binary() {
    info "Building radb-client binary..."

    # Get version information
    if [ -f VERSION ]; then
        VERSION=$(cat VERSION)
    else
        VERSION="dev"
    fi

    if command -v git &> /dev/null && [ -d .git ]; then
        GIT_COMMIT=$(git rev-parse --short HEAD 2>/dev/null || echo "unknown")
        GIT_BRANCH=$(git branch --show-current 2>/dev/null || echo "unknown")
    else
        GIT_COMMIT="unknown"
        GIT_BRANCH="unknown"
    fi

    BUILD_DATE=$(date -u '+%Y-%m-%d_%H:%M:%S_UTC')

    # Build ldflags
    LDFLAGS="-s -w"
    LDFLAGS="$LDFLAGS -X 'github.com/bss/radb-client/internal/version.Version=$VERSION'"
    LDFLAGS="$LDFLAGS -X 'github.com/bss/radb-client/internal/version.GitCommit=$GIT_COMMIT'"
    LDFLAGS="$LDFLAGS -X 'github.com/bss/radb-client/internal/version.GitBranch=$GIT_BRANCH'"
    LDFLAGS="$LDFLAGS -X 'github.com/bss/radb-client/internal/version.BuildDate=$BUILD_DATE'"

    # Create bin directory
    mkdir -p bin

    # Build
    info "Building for ${OS}/${ARCH}..."
    if ! go build -ldflags "$LDFLAGS" -o bin/radb-client ./cmd/radb-client; then
        error "Build failed"
        exit 1
    fi

    success "Binary built successfully: bin/radb-client"
}

# Prompt for installation type
prompt_install_type() {
    echo ""
    info "Installation Options:"
    echo "  1) User installation (\$HOME/bin) - no sudo required"
    echo "  2) System installation (/usr/local/bin) - requires sudo"
    echo "  3) Custom location"
    echo "  4) Skip installation (just build)"
    echo ""

    while true; do
        read -p "Choose installation type [1-4]: " choice
        case $choice in
            1)
                INSTALL_DIR="$HOME/bin"
                NEED_SUDO=false
                break
                ;;
            2)
                INSTALL_DIR="/usr/local/bin"
                NEED_SUDO=true
                break
                ;;
            3)
                read -p "Enter custom installation directory: " INSTALL_DIR
                # Expand tilde
                INSTALL_DIR="${INSTALL_DIR/#\~/$HOME}"
                if [[ "$INSTALL_DIR" == /usr/* ]] || [[ "$INSTALL_DIR" == /opt/* ]]; then
                    NEED_SUDO=true
                else
                    NEED_SUDO=false
                fi
                break
                ;;
            4)
                INSTALL_DIR=""
                info "Skipping installation. Binary is at: bin/radb-client"
                return
                ;;
            *)
                error "Invalid choice. Please enter 1-4."
                ;;
        esac
    done
}

# Install binary
install_binary() {
    if [ -z "$INSTALL_DIR" ]; then
        return
    fi

    info "Installing to: $INSTALL_DIR"

    # Create directory if needed
    if [ ! -d "$INSTALL_DIR" ]; then
        if [ "$NEED_SUDO" = true ]; then
            sudo mkdir -p "$INSTALL_DIR"
        else
            mkdir -p "$INSTALL_DIR"
        fi
    fi

    # Copy binary
    if [ "$NEED_SUDO" = true ]; then
        sudo cp bin/radb-client "$INSTALL_DIR/radb-client"
        sudo chmod +x "$INSTALL_DIR/radb-client"
    else
        cp bin/radb-client "$INSTALL_DIR/radb-client"
        chmod +x "$INSTALL_DIR/radb-client"
    fi

    success "Binary installed to: $INSTALL_DIR/radb-client"

    # Check if directory is in PATH
    if [[ ":$PATH:" != *":$INSTALL_DIR:"* ]]; then
        warning "Directory $INSTALL_DIR is not in your PATH"
        echo ""
        info "Add this line to your ~/.bashrc or ~/.zshrc:"
        echo "    export PATH=\"\$PATH:$INSTALL_DIR\""
        echo ""
    fi
}

# Initialize configuration
prompt_config_setup() {
    echo ""
    read -p "Would you like to initialize configuration now? [Y/n]: " response
    case "$response" in
        [nN]|[nN][oO])
            info "Skipping configuration. Run 'radb-client config init' later."
            return
            ;;
    esac

    info "Initializing configuration..."

    # Determine which binary to use
    if [ -n "$INSTALL_DIR" ] && [ -f "$INSTALL_DIR/radb-client" ]; then
        BINARY="$INSTALL_DIR/radb-client"
    else
        BINARY="./bin/radb-client"
    fi

    # Initialize config
    if ! $BINARY config init; then
        error "Failed to initialize configuration"
        return
    fi

    success "Configuration initialized"
}

# Setup credentials
prompt_credentials_setup() {
    echo ""
    read -p "Would you like to configure RADb credentials now? [Y/n]: " response
    case "$response" in
        [nN]|[nN][oO])
            info "Skipping credentials. Run 'radb-client auth login' later."
            return
            ;;
    esac

    # Determine which binary to use
    if [ -n "$INSTALL_DIR" ] && [ -f "$INSTALL_DIR/radb-client" ]; then
        BINARY="$INSTALL_DIR/radb-client"
    else
        BINARY="./bin/radb-client"
    fi

    info "Setting up RADb credentials..."
    echo ""
    info "You will be prompted for your RADb username and password."
    info "Your password will be encrypted and stored securely."
    echo ""

    if ! $BINARY auth login; then
        warning "Credential setup skipped or failed. Run 'radb-client auth login' later."
        return
    fi

    success "Credentials configured successfully"
}

# Prompt for daemon installation
prompt_daemon_setup() {
    echo ""
    read -p "Would you like to install radb-client as a systemd daemon? [y/N]: " response
    case "$response" in
        [yY]|[yY][eE][sS])
            ;;
        *)
            info "Skipping daemon installation."
            info "Run './scripts/install-daemon.sh' later if you want daemon mode."
            return
            ;;
    esac

    if [ ! -f "scripts/install-daemon.sh" ]; then
        error "Daemon install script not found: scripts/install-daemon.sh"
        return
    fi

    info "Running daemon installation..."
    echo ""

    # Check if we need to set environment variables for daemon
    read -p "Would you like to configure credentials for the daemon? [y/N]: " cred_response
    case "$cred_response" in
        [yY]|[yY][eE][sS])
            echo ""
            read -p "Enter RADb username: " RADB_USER
            read -s -p "Enter RADb password: " RADB_PASS
            echo ""

            export RADB_USERNAME="$RADB_USER"
            export RADB_PASSWORD="$RADB_PASS"
            ;;
    esac

    if ! sudo bash scripts/install-daemon.sh; then
        error "Daemon installation failed"
        return
    fi

    success "Daemon installed successfully"
}

# Display next steps
show_next_steps() {
    echo ""
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
    success "Installation Complete!"
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
    echo ""

    # Show binary location
    if [ -n "$INSTALL_DIR" ]; then
        info "Binary installed at: $INSTALL_DIR/radb-client"
    else
        info "Binary built at: bin/radb-client"
    fi

    # Show version
    if [ -n "$INSTALL_DIR" ] && [ -f "$INSTALL_DIR/radb-client" ]; then
        BINARY="$INSTALL_DIR/radb-client"
    else
        BINARY="./bin/radb-client"
    fi

    VERSION_OUTPUT=$($BINARY version --short 2>/dev/null || echo "unknown")
    info "Version: $VERSION_OUTPUT"

    echo ""
    info "Next Steps:"
    echo ""

    # Configuration
    if [ ! -f "$HOME/.radb-client/config.yaml" ]; then
        echo "  1. Initialize configuration:"
        echo "     radb-client config init"
        echo ""
    fi

    # Authentication
    echo "  2. Authenticate with RADb:"
    echo "     radb-client auth login"
    echo ""

    # Basic usage
    echo "  3. Try these commands:"
    echo "     radb-client route list          # List your routes"
    echo "     radb-client contact list        # List your contacts"
    echo "     radb-client snapshot create     # Create a snapshot"
    echo ""

    # Documentation
    echo "  4. Learn more:"
    echo "     radb-client --help"
    echo "     radb-client route --help"
    echo "     cat QUICKSTART.md"
    echo ""

    # Testing
    if [ -f "TESTING_RUNBOOK.md" ]; then
        echo "  5. Manual testing:"
        echo "     See TESTING_RUNBOOK.md for comprehensive test procedures"
        echo ""
    fi

    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
}

# Main installation flow
main() {
    echo ""
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
    echo "       RADb Client - Interactive Installation"
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
    echo ""

    # Check if we're in the right directory
    if [ ! -f "go.mod" ] || [ ! -d "cmd/radb-client" ]; then
        error "This script must be run from the radb-client project root"
        exit 1
    fi

    # Run installation steps
    detect_platform
    check_prerequisites
    build_binary
    prompt_install_type
    install_binary
    prompt_config_setup
    prompt_credentials_setup

    # Only offer daemon on Linux
    if [ "$OS" = "linux" ]; then
        prompt_daemon_setup
    fi

    show_next_steps

    echo ""
    success "All done! 🚀"
    echo ""
}

# Run main
main "$@"
