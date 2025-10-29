#!/bin/bash
#
# RADb Client Daemon Installation Script
# For Ubuntu 22.04 LTS and later
#
# This script installs radb-client as a systemd service that runs as a daemon,
# automatically monitoring RADb for changes and maintaining historical snapshots.
#
# Usage: sudo ./install-daemon.sh [options]
#

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Default values
INSTALL_USER="${RADB_USER:-radb}"
INSTALL_DIR="/opt/radb-client"
CONFIG_DIR="/etc/radb-client"
DATA_DIR="/var/lib/radb-client"
LOG_DIR="/var/log/radb-client"
BINARY_PATH=""
CHECK_INTERVAL="3600"  # Default: check every hour (in seconds)
ENABLE_SERVICE="yes"
START_SERVICE="yes"

# Script information
SCRIPT_VERSION="0.9.0"
MIN_UBUNTU_VERSION="22.04"

#######################################
# Print functions
#######################################

print_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

print_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

print_header() {
    echo ""
    echo "=========================================="
    echo "$1"
    echo "=========================================="
    echo ""
}

#######################################
# Usage
#######################################

usage() {
    cat << EOF
RADb Client Daemon Installation Script v${SCRIPT_VERSION}

Usage: sudo $0 [OPTIONS]

Options:
    -b, --binary PATH       Path to radb-client binary (required)
    -u, --user USER         User to run daemon as (default: radb)
    -i, --interval SECONDS  Check interval in seconds (default: 3600)
    -c, --config-dir DIR    Configuration directory (default: /etc/radb-client)
    -d, --data-dir DIR      Data directory (default: /var/lib/radb-client)
    -l, --log-dir DIR       Log directory (default: /var/log/radb-client)
    --no-enable             Don't enable service on boot
    --no-start              Don't start service after installation
    -h, --help              Show this help message

Examples:
    # Install with default settings
    sudo $0 --binary ./dist/radb-client

    # Install with custom user and check interval (every 30 minutes)
    sudo $0 --binary ./dist/radb-client --user myuser --interval 1800

    # Install without starting immediately
    sudo $0 --binary ./dist/radb-client --no-start

Environment Variables:
    RADB_USER               Default user for daemon (overridden by --user)
    RADB_API_USERNAME       RADb API username for configuration
    RADB_API_KEY            RADb API key for configuration

EOF
    exit 1
}

#######################################
# Parse arguments
#######################################

parse_args() {
    while [[ $# -gt 0 ]]; do
        case $1 in
            -b|--binary)
                BINARY_PATH="$2"
                shift 2
                ;;
            -u|--user)
                INSTALL_USER="$2"
                shift 2
                ;;
            -i|--interval)
                CHECK_INTERVAL="$2"
                shift 2
                ;;
            -c|--config-dir)
                CONFIG_DIR="$2"
                shift 2
                ;;
            -d|--data-dir)
                DATA_DIR="$2"
                shift 2
                ;;
            -l|--log-dir)
                LOG_DIR="$2"
                shift 2
                ;;
            --no-enable)
                ENABLE_SERVICE="no"
                shift
                ;;
            --no-start)
                START_SERVICE="no"
                shift
                ;;
            -h|--help)
                usage
                ;;
            *)
                print_error "Unknown option: $1"
                usage
                ;;
        esac
    done

    # Validate required arguments
    if [[ -z "$BINARY_PATH" ]]; then
        print_error "Binary path is required. Use --binary option."
        usage
    fi

    if [[ ! -f "$BINARY_PATH" ]]; then
        print_error "Binary not found at: $BINARY_PATH"
        exit 1
    fi
}

#######################################
# System checks
#######################################

check_system() {
    print_header "System Checks"

    # Check if running as root
    if [[ $EUID -ne 0 ]]; then
        print_error "This script must be run as root (use sudo)"
        exit 1
    fi
    print_success "Running as root"

    # Check Ubuntu version
    if [[ -f /etc/os-release ]]; then
        source /etc/os-release
        if [[ "$ID" != "ubuntu" ]]; then
            print_warning "This script is designed for Ubuntu. Current OS: $ID"
            read -p "Continue anyway? [y/N] " -n 1 -r
            echo
            if [[ ! $REPLY =~ ^[Yy]$ ]]; then
                exit 1
            fi
        fi

        # Parse version
        UBUNTU_VERSION="${VERSION_ID}"
        print_info "Detected Ubuntu ${UBUNTU_VERSION}"

        # Check minimum version (22.04)
        if [[ $(echo -e "${UBUNTU_VERSION}\n${MIN_UBUNTU_VERSION}" | sort -V | head -n1) != "${MIN_UBUNTU_VERSION}" ]]; then
            print_warning "Ubuntu version ${UBUNTU_VERSION} is older than recommended ${MIN_UBUNTU_VERSION}"
        else
            print_success "Ubuntu version ${UBUNTU_VERSION} is supported"
        fi
    else
        print_warning "Cannot detect Ubuntu version"
    fi

    # Check systemd
    if ! command -v systemctl &> /dev/null; then
        print_error "systemd not found. This script requires systemd."
        exit 1
    fi
    print_success "systemd detected"

    # Check binary
    if [[ ! -x "$BINARY_PATH" ]]; then
        print_warning "Binary is not executable. Making it executable..."
        chmod +x "$BINARY_PATH"
    fi
    print_success "Binary is executable"

    # Test binary
    if ! "$BINARY_PATH" version &> /dev/null; then
        print_error "Binary test failed. The binary may be corrupted."
        exit 1
    fi
    print_success "Binary test passed"
}

#######################################
# Create user
#######################################

create_user() {
    print_header "User Setup"

    if id "$INSTALL_USER" &>/dev/null; then
        print_info "User '$INSTALL_USER' already exists"
    else
        print_info "Creating system user '$INSTALL_USER'..."
        useradd --system --no-create-home --shell /usr/sbin/nologin "$INSTALL_USER"
        print_success "User '$INSTALL_USER' created"
    fi
}

#######################################
# Create directories
#######################################

create_directories() {
    print_header "Directory Setup"

    # Create directories
    for dir in "$INSTALL_DIR" "$CONFIG_DIR" "$DATA_DIR" "$DATA_DIR/cache" "$DATA_DIR/history" "$LOG_DIR"; do
        if [[ ! -d "$dir" ]]; then
            print_info "Creating directory: $dir"
            mkdir -p "$dir"
        else
            print_info "Directory exists: $dir"
        fi
    done

    # Set permissions
    print_info "Setting permissions..."
    chown -R "$INSTALL_USER:$INSTALL_USER" "$DATA_DIR"
    chown -R "$INSTALL_USER:$INSTALL_USER" "$LOG_DIR"
    chown root:root "$CONFIG_DIR"
    chmod 755 "$INSTALL_DIR"
    chmod 755 "$CONFIG_DIR"
    chmod 750 "$DATA_DIR"
    chmod 750 "$LOG_DIR"

    print_success "Directories created and permissions set"
}

#######################################
# Install binary
#######################################

install_binary() {
    print_header "Binary Installation"

    print_info "Copying binary to $INSTALL_DIR..."
    cp "$BINARY_PATH" "$INSTALL_DIR/radb-client"
    chown root:root "$INSTALL_DIR/radb-client"
    chmod 755 "$INSTALL_DIR/radb-client"

    # Create symlink in /usr/local/bin for easy access
    print_info "Creating symlink in /usr/local/bin..."
    ln -sf "$INSTALL_DIR/radb-client" /usr/local/bin/radb-client

    print_success "Binary installed"
}

#######################################
# Create configuration
#######################################

create_configuration() {
    print_header "Configuration Setup"

    local config_file="$CONFIG_DIR/config.yaml"

    if [[ -f "$config_file" ]]; then
        print_warning "Configuration file exists. Creating backup..."
        cp "$config_file" "$config_file.backup.$(date +%s)"
    fi

    print_info "Creating configuration file: $config_file"

    cat > "$config_file" << EOF
# RADb Client Configuration
# Daemon Mode Configuration

api:
  base_url: https://api.radb.net
  source: RADB
  format: json
  timeout: 30

  rate_limit:
    requests_per_minute: 60
    burst_size: 10

  retry:
    max_attempts: 3
    backoff_multiplier: 2
    initial_delay_ms: 1000

preferences:
  # Data directories (overridden by daemon)
  cache_dir: ${DATA_DIR}/cache
  history_dir: ${DATA_DIR}/history

  # Logging
  log_level: INFO

  # Snapshot retention
  max_snapshots: 100
  auto_snapshot: true

  # Output format
  output_format: json
  color: false  # Disabled for daemon mode

daemon:
  # Check interval in seconds
  check_interval: ${CHECK_INTERVAL}

  # Enable change notifications (future feature)
  notify_on_changes: false

  # Cleanup old snapshots automatically
  auto_cleanup: true

  # Snapshot retention for daemon mode
  retention:
    # Keep snapshots for 90 days
    max_age_days: 90
    # Keep at least this many recent snapshots
    min_snapshots: 10
EOF

    chmod 640 "$config_file"
    chown root:"$INSTALL_USER" "$config_file"

    print_success "Configuration file created"

    # Check for credentials in environment
    if [[ -n "$RADB_API_USERNAME" ]] && [[ -n "$RADB_API_KEY" ]]; then
        print_info "Configuring credentials from environment variables..."
        configure_credentials
    else
        print_warning "No credentials found in environment variables (RADB_API_USERNAME, RADB_API_KEY)"
        print_info "You'll need to configure credentials manually before starting the service:"
        print_info "  sudo -u $INSTALL_USER radb-client auth login"
    fi
}

#######################################
# Configure credentials
#######################################

configure_credentials() {
    # Create a temporary script to run as the daemon user
    local temp_script="/tmp/radb-configure-$$"

    cat > "$temp_script" << EOF
#!/bin/bash
export HOME="$DATA_DIR"
export XDG_CONFIG_HOME="$CONFIG_DIR"

# Configure using environment variables
echo "$RADB_API_KEY" | radb-client auth login --username "$RADB_API_USERNAME" --password-stdin
EOF

    chmod 700 "$temp_script"

    # Run as daemon user
    if su - "$INSTALL_USER" -s /bin/bash -c "$temp_script" 2>&1; then
        print_success "Credentials configured"
    else
        print_error "Failed to configure credentials"
    fi

    rm -f "$temp_script"
}

#######################################
# Create systemd service
#######################################

create_systemd_service() {
    print_header "Systemd Service Setup"

    local service_file="/etc/systemd/system/radb-client.service"

    print_info "Creating systemd service file: $service_file"

    cat > "$service_file" << EOF
[Unit]
Description=RADb Client Daemon - Route Object Monitoring
Documentation=https://github.com/bss/radb-client
After=network-online.target
Wants=network-online.target

[Service]
Type=simple
User=${INSTALL_USER}
Group=${INSTALL_USER}

# Environment
Environment="HOME=${DATA_DIR}"
Environment="XDG_CONFIG_HOME=${CONFIG_DIR}"

# Execution
ExecStart=${INSTALL_DIR}/radb-client daemon --config ${CONFIG_DIR}/config.yaml
ExecReload=/bin/kill -HUP \$MAINPID

# Working directory
WorkingDirectory=${DATA_DIR}

# Restart policy
Restart=on-failure
RestartSec=30s

# Resource limits
LimitNOFILE=65536

# Logging
StandardOutput=append:${LOG_DIR}/radb-client.log
StandardError=append:${LOG_DIR}/radb-client-error.log
SyslogIdentifier=radb-client

# Security hardening
NoNewPrivileges=true
PrivateTmp=true
ProtectSystem=strict
ProtectHome=true
ReadWritePaths=${DATA_DIR} ${LOG_DIR}
ReadOnlyPaths=${CONFIG_DIR}

# Additional security
ProtectKernelTunables=true
ProtectKernelModules=true
ProtectControlGroups=true
RestrictRealtime=true
RestrictNamespaces=true

[Install]
WantedBy=multi-user.target
EOF

    chmod 644 "$service_file"

    print_success "Systemd service file created"

    # Reload systemd
    print_info "Reloading systemd daemon..."
    systemctl daemon-reload
    print_success "Systemd daemon reloaded"
}

#######################################
# Create timer (alternative to daemon mode)
#######################################

create_systemd_timer() {
    print_header "Systemd Timer Setup (Alternative)"

    print_info "Creating systemd timer for periodic execution..."

    # Create service for one-shot execution
    cat > "/etc/systemd/system/radb-client-sync.service" << EOF
[Unit]
Description=RADb Client Sync - One-time Route Check
After=network-online.target

[Service]
Type=oneshot
User=${INSTALL_USER}
Group=${INSTALL_USER}

Environment="HOME=${DATA_DIR}"
Environment="XDG_CONFIG_HOME=${CONFIG_DIR}"

ExecStart=${INSTALL_DIR}/radb-client route list
ExecStartPost=${INSTALL_DIR}/radb-client route diff
ExecStartPost=${INSTALL_DIR}/radb-client snapshot create

WorkingDirectory=${DATA_DIR}

StandardOutput=append:${LOG_DIR}/radb-sync.log
StandardError=append:${LOG_DIR}/radb-sync-error.log
EOF

    # Create timer
    cat > "/etc/systemd/system/radb-client-sync.timer" << EOF
[Unit]
Description=RADb Client Sync Timer
Documentation=https://github.com/bss/radb-client

[Timer]
# Run on boot after 5 minutes
OnBootSec=5min

# Run every hour
OnUnitActiveSec=${CHECK_INTERVAL}s

# Persistent across reboots
Persistent=true

[Install]
WantedBy=timers.target
EOF

    chmod 644 /etc/systemd/system/radb-client-sync.*

    systemctl daemon-reload

    print_success "Systemd timer created"
    print_info "To use timer instead of daemon:"
    print_info "  sudo systemctl disable radb-client.service"
    print_info "  sudo systemctl enable --now radb-client-sync.timer"
}

#######################################
# Create logrotate configuration
#######################################

create_logrotate() {
    print_header "Logrotate Setup"

    local logrotate_file="/etc/logrotate.d/radb-client"

    print_info "Creating logrotate configuration: $logrotate_file"

    cat > "$logrotate_file" << EOF
${LOG_DIR}/*.log {
    daily
    rotate 30
    compress
    delaycompress
    notifempty
    missingok
    create 0640 ${INSTALL_USER} ${INSTALL_USER}
    sharedscripts
    postrotate
        systemctl reload radb-client.service > /dev/null 2>&1 || true
    endscript
}
EOF

    chmod 644 "$logrotate_file"

    print_success "Logrotate configuration created"
}

#######################################
# Enable and start service
#######################################

enable_service() {
    print_header "Service Activation"

    if [[ "$ENABLE_SERVICE" == "yes" ]]; then
        print_info "Enabling service to start on boot..."
        systemctl enable radb-client.service
        print_success "Service enabled"
    else
        print_info "Service not enabled (use --no-enable flag)"
    fi

    if [[ "$START_SERVICE" == "yes" ]]; then
        print_info "Starting service..."
        if systemctl start radb-client.service; then
            print_success "Service started"
            sleep 2
            print_info "Service status:"
            systemctl status radb-client.service --no-pager -l || true
        else
            print_error "Failed to start service"
            print_info "Check logs: journalctl -u radb-client.service -n 50"
        fi
    else
        print_info "Service not started (use --no-start flag)"
    fi
}

#######################################
# Create helper scripts
#######################################

create_helper_scripts() {
    print_header "Helper Scripts"

    # Create management script
    local mgmt_script="/usr/local/bin/radb-daemon"

    print_info "Creating management helper: $mgmt_script"

    cat > "$mgmt_script" << 'EOF'
#!/bin/bash
# RADb Client Daemon Management Helper

case "$1" in
    status)
        systemctl status radb-client.service
        ;;
    logs)
        journalctl -u radb-client.service -f
        ;;
    start)
        sudo systemctl start radb-client.service
        ;;
    stop)
        sudo systemctl stop radb-client.service
        ;;
    restart)
        sudo systemctl restart radb-client.service
        ;;
    enable)
        sudo systemctl enable radb-client.service
        ;;
    disable)
        sudo systemctl disable radb-client.service
        ;;
    diff)
        sudo -u radb radb-client route diff
        ;;
    snapshots)
        sudo -u radb radb-client snapshot list
        ;;
    history)
        sudo -u radb radb-client history show "$@"
        ;;
    *)
        echo "RADb Client Daemon Management"
        echo ""
        echo "Usage: radb-daemon <command>"
        echo ""
        echo "Commands:"
        echo "  status      - Show service status"
        echo "  logs        - Tail service logs"
        echo "  start       - Start service"
        echo "  stop        - Stop service"
        echo "  restart     - Restart service"
        echo "  enable      - Enable service on boot"
        echo "  disable     - Disable service on boot"
        echo "  diff        - Show route changes"
        echo "  snapshots   - List snapshots"
        echo "  history     - Show change history"
        ;;
esac
EOF

    chmod 755 "$mgmt_script"

    print_success "Helper scripts created"
}

#######################################
# Print summary
#######################################

print_summary() {
    print_header "Installation Complete!"

    cat << EOF
${GREEN}RADb Client daemon has been successfully installed!${NC}

Installation Details:
  User:            ${INSTALL_USER}
  Install Dir:     ${INSTALL_DIR}
  Config Dir:      ${CONFIG_DIR}
  Data Dir:        ${DATA_DIR}
  Log Dir:         ${LOG_DIR}
  Check Interval:  ${CHECK_INTERVAL} seconds ($(($CHECK_INTERVAL / 60)) minutes)

Service Management:
  Start service:    sudo systemctl start radb-client
  Stop service:     sudo systemctl stop radb-client
  Restart service:  sudo systemctl restart radb-client
  Service status:   sudo systemctl status radb-client
  View logs:        sudo journalctl -u radb-client -f

Quick Helper Commands:
  radb-daemon status      - Show service status
  radb-daemon logs        - Tail logs in real-time
  radb-daemon diff        - Show recent route changes
  radb-daemon snapshots   - List all snapshots
  radb-daemon history     - Show change history

Manual Commands (as radb user):
  sudo -u ${INSTALL_USER} radb-client route list
  sudo -u ${INSTALL_USER} radb-client route diff
  sudo -u ${INSTALL_USER} radb-client snapshot list
  sudo -u ${INSTALL_USER} radb-client history show

Log Files:
  Service log:  ${LOG_DIR}/radb-client.log
  Error log:    ${LOG_DIR}/radb-client-error.log
  Journalctl:   journalctl -u radb-client

Configuration:
  Config file:  ${CONFIG_DIR}/config.yaml
  Edit config:  sudo nano ${CONFIG_DIR}/config.yaml
  After edit:   sudo systemctl restart radb-client

Data:
  Snapshots:    ${DATA_DIR}/cache/
  History:      ${DATA_DIR}/history/

Alternative: Systemd Timer
  If you prefer periodic execution instead of a long-running daemon:
    sudo systemctl disable radb-client.service
    sudo systemctl enable --now radb-client-sync.timer
    sudo systemctl status radb-client-sync.timer

Next Steps:
  1. Configure credentials (if not already done):
     sudo -u ${INSTALL_USER} radb-client auth login

  2. Test the service:
     sudo systemctl status radb-client

  3. View initial logs:
     sudo journalctl -u radb-client -n 100

  4. Check for changes:
     radb-daemon diff

For more information:
  - Documentation: /opt/radb-client/docs/
  - Service file: /etc/systemd/system/radb-client.service
  - Helper: radb-daemon --help

${GREEN}Happy monitoring!${NC}
EOF
}

#######################################
# Main installation flow
#######################################

main() {
    print_header "RADb Client Daemon Installation"
    print_info "Version: $SCRIPT_VERSION"
    print_info "Target: Ubuntu ${MIN_UBUNTU_VERSION}+"

    parse_args "$@"
    check_system
    create_user
    create_directories
    install_binary
    create_configuration
    create_systemd_service
    create_systemd_timer
    create_logrotate
    create_helper_scripts

    if [[ "$ENABLE_SERVICE" == "yes" ]] || [[ "$START_SERVICE" == "yes" ]]; then
        enable_service
    fi

    print_summary
}

# Run main function
main "$@"
