#!/bin/bash

set -e

# Цвета для вывода
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Функция для вывода цветного текста
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

# Проверка root прав
if [ "$EUID" -ne 0 ]; then
  print_error "Please run the script with root privileges (sudo)."
  exit 1
fi

# Определение архитектуры
ARCH=""
if [ "$(uname -m)" == "x86_64" ]; then
  ARCH="amd64"
elif [ "$(uname -m)" == "aarch64" ]; then
  ARCH="arm64"
else
  print_error "Unsupported architecture: $(uname -m)"
  exit 1
fi

# Определение дистрибутива и пакетного менеджера
DISTRO=""
PKG_MANAGER=""
INSTALL_FROM_PACKAGE=false

if command -v apt-get &> /dev/null; then
    DISTRO="debian"
    PKG_MANAGER="apt-get"
    INSTALL_FROM_PACKAGE=true
elif command -v yum &> /dev/null; then
    DISTRO="rhel"
    PKG_MANAGER="yum"
    INSTALL_FROM_PACKAGE=true
elif command -v dnf &> /dev/null; then
    DISTRO="fedora"
    PKG_MANAGER="dnf"
    INSTALL_FROM_PACKAGE=true
elif command -v pacman &> /dev/null; then
    DISTRO="arch"
    PKG_MANAGER="pacman"
    INSTALL_FROM_PACKAGE=false
else
    print_warning "Unable to determine package manager. Will install from releases."
    INSTALL_FROM_PACKAGE=false
fi

print_info "Detected distribution: $DISTRO ($PKG_MANAGER)"

# Остановка существующего сервиса
if systemctl is-active --quiet tblocker; then
  print_info "Stopping existing tblocker service..."
  systemctl stop tblocker
fi

# Установка из пакетов
if [ "$INSTALL_FROM_PACKAGE" = true ]; then
    print_info "Installing from package repository..."
    
    case $PKG_MANAGER in
        "apt-get")
            print_info "Adding xray-tools repository..."
            apt-get update -qq > /dev/null
            apt-get install -y curl gnupg > /dev/null
            curl -s https://repo.remna.dev/xray-tools/public.gpg | gpg --yes --dearmor -o /usr/share/keyrings/openrepo-xray-tools.gpg > /dev/null
            echo "deb [arch=any signed-by=/usr/share/keyrings/openrepo-xray-tools.gpg] https://repo.remna.dev/xray-tools/ stable main" > /etc/apt/sources.list.d/openrepo-xray-tools.list
            apt-get update -qq > /dev/null
            apt-get install -y tblocker > /dev/null
            ;;
        "yum")
            print_info "Adding xray-tools repository..."
            echo """
[xray-tools-rpm]
name=xray-tools-rpm
baseurl=https://repo.remna.dev/xray-tools-rpm
enabled=1
repo_gpgcheck=1
gpgkey=https://repo.remna.dev/xray-tools-rpm/public.gpg
""" > /etc/yum.repos.d/xray-tools-rpm.repo
            yum install -y tblocker
            ;;
    esac
    
    if [ $? -eq 0 ]; then
        print_success "Successfully installed from package repository"
        INSTALL_DIR="/opt/tblocker"
        CONFIG_PATH="/opt/tblocker/config.yaml"
    else
        print_warning "Failed to install from package repository, falling back to releases"
        INSTALL_FROM_PACKAGE=false
    fi
fi

# Установка из релизов
if [ "$INSTALL_FROM_PACKAGE" = false ]; then
    print_info "Installing from GitHub releases..."
    
    # Установка зависимостей
    print_info "Installing necessary dependencies..."
    case $PKG_MANAGER in
        "apt-get")
            apt-get update -qq
            apt-get install -y curl conntrack-tools > /dev/null
            ;;
        "yum"|"dnf")
            if [ "$PKG_MANAGER" = "yum" ]; then
                yum install -y epel-release > /dev/null
                yum install -y curl conntrack-tools > /dev/null
            else
                dnf install -y curl conntrack-tools > /dev/null
            fi
            ;;
        "pacman")
            pacman -Sy --noconfirm curl conntrack-tools > /dev/null
            ;;
        *)
            print_warning "Please install curl manually if not already installed."
            ;;
    esac
    
    # Загрузка последней версии
    print_info "Downloading the latest version of tblocker..."
    LATEST_RELEASE=$(curl -s https://api.github.com/repos/kutovoys/xray-torrent-blocker/releases/latest | grep tag_name | cut -d '"' -f 4)
    URL="https://github.com/kutovoys/xray-torrent-blocker/releases/download/${LATEST_RELEASE}/xray-torrent-blocker-${LATEST_RELEASE}-linux-${ARCH}.tar.gz"
    
    curl -sL "$URL" -o tblocker.tar.gz
    
    if [ ! -f "tblocker.tar.gz" ]; then
        print_error "Failed to download tblocker"
        exit 1
    fi
    
    # Распаковка
    print_info "Extracting files..."
    mkdir -p /opt/tblocker
    tar -xzf tblocker.tar.gz -C /opt/tblocker --overwrite
    rm tblocker.tar.gz
    
    INSTALL_DIR="/opt/tblocker"
    CONFIG_PATH="/opt/tblocker/config.yaml"
    CONFIG_TEMPLATE_PATH="/opt/tblocker/config.yaml.default"
    
    print_info "Loading nf_conntrack kernel module..."
    modprobe nf_conntrack || true
    echo 'nf_conntrack' > /etc/modules-load.d/conntrack.conf
    print_success "nf_conntrack kernel module loaded"

    # Копирование конфига
    if [ ! -f "$CONFIG_PATH" ]; then
        cp "$CONFIG_TEMPLATE_PATH" "$CONFIG_PATH"
        print_info "New configuration file created at $CONFIG_PATH"
    else
        print_info "Configuration file already exists at $CONFIG_PATH"
    fi
    
    # Копирование systemd сервиса
    print_info "Setting up systemd service..."
    curl -sL https://raw.githubusercontent.com/kutovoys/xray-torrent-blocker/main/tblocker.service -o /etc/systemd/system/tblocker.service
fi

# Запрос пути к файлу логов
print_info "Configuration setup..."
echo ""
read -p "Enter the path to the log file to monitor: " log_file_path

# Проверка существования файла логов
if [ ! -f "$log_file_path" ]; then
    print_warning "Log file does not exist: $log_file_path"
    read -p "Do you want to create it? (y/N): " create_log_file
    if [[ $create_log_file =~ ^[Yy]$ ]]; then
        mkdir -p "$(dirname "$log_file_path")"
        touch "$log_file_path"
        print_info "Created log file: $log_file_path"
    fi
fi

# Выбор файрвола
echo ""
print_info "Available firewalls:"
echo "1) iptables (Linux netfilter)"
echo "2) nft (nftables)"
echo ""

while true; do
    read -p "Select firewall (1-2): " firewall_choice
    case $firewall_choice in
        1) FIREWALL="iptables"; break ;;
        2) FIREWALL="nft"; break ;;
        *) print_error "Invalid choice. Please select 1 or 2." ;;
    esac
done

# Проверка и установка выбранного файрвола
print_info "Checking firewall availability..."
case $FIREWALL in
    "iptables")
        if ! command -v iptables &> /dev/null; then
            print_info "Installing iptables..."
            case $PKG_MANAGER in
                "apt-get")
                    apt-get install -y iptables
                    ;;
                "yum"|"dnf")
                    if [ "$PKG_MANAGER" = "yum" ]; then
                        yum install -y iptables-services
                    else
                        dnf install -y iptables-services
                    fi
                    ;;
                "pacman")
                    pacman -S --noconfirm iptables
                    ;;
            esac
        fi
        ;;
    "nft")
        if ! command -v nft &> /dev/null; then
            print_info "Installing nftables..."
            case $PKG_MANAGER in
                "apt-get")
                    apt-get install -y nftables
                    ;;
                "yum"|"dnf")
                    if [ "$PKG_MANAGER" = "yum" ]; then
                        yum install -y nftables
                    else
                        dnf install -y nftables
                    fi
                    ;;
                "pacman")
                    pacman -S --noconfirm nftables
                    ;;
            esac
        fi
        ;;
esac

# Обновление конфигурации
print_info "Updating configuration..."
sed -i "s|LogFile: \".*\"|LogFile: \"$log_file_path\"|" "$CONFIG_PATH"
sed -i "s|BlockMode: \".*\"|BlockMode: \"$FIREWALL\"|" "$CONFIG_PATH"

print_success "Configuration updated:"
print_info "  Log file: $log_file_path"
print_info "  Firewall: $FIREWALL"

# Запуск сервиса
print_info "Starting tblocker service..."
systemctl daemon-reload
systemctl enable tblocker
systemctl start tblocker

# Проверка статуса
if systemctl is-active --quiet tblocker; then
    print_success "tblocker service is running successfully!"
else
    print_error "Failed to start tblocker service"
    systemctl status tblocker --no-pager
    exit 1
fi

echo ""
echo "==============================================================="
print_success "Installation complete! The tblocker service is running."
echo "==============================================================="
echo ""
print_info "Configuration file: $CONFIG_PATH"
print_info "Service status: systemctl status tblocker"
print_info "Service logs: journalctl -u tblocker -f"
echo ""
print_warning "IMPORTANT: Make sure your firewall ($FIREWALL) is properly configured!"
echo ""
print_info "For additional parameters (webhooks, whitelist, etc.) - see documentation:"
print_info "https://github.com/kutovoys/xray-torrent-blocker"
echo "==============================================================="