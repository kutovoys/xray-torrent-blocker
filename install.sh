#!/bin/bash

set -e

if [ "$EUID" -ne 0 ]; then
  echo "Please run the script with root privileges (sudo)."
  exit 1
fi

echo "Installing necessary dependencies..."
if command -v apt-get &> /dev/null; then
  apt-get update -qq
  apt-get install -y ufw curl > /dev/null
elif command -v yum &> /dev/null; then
  yum install -y epel-release > /dev/null
  yum install -y ufw curl > /dev/null
elif command -v dnf &> /dev/null; then
  dnf install -y ufw curl > /dev/null
elif command -v pacman &> /dev/null; then
  pacman -Sy --noconfirm ufw curl > /dev/null
else
  echo "Unable to determine package manager. Please install ufw and curl manually."
  exit 1
fi

if systemctl is-active --quiet torrent-blocker; then
  echo "Stopping existing torrent-blocker service..."
  systemctl stop torrent-blocker
fi

ARCH=""
if [ "$(uname -m)" == "x86_64" ]; then
  ARCH="amd64"
elif [ "$(uname -m)" == "aarch64" ];then
  ARCH="arm64"
else
  echo "Unsupported architecture."
  exit 1
fi

echo "Downloading the latest version of torrent-blocker..."
LATEST_RELEASE=$(curl -s https://api.github.com/repos/kutovoys/xray-torrent-blocker/releases/latest | grep tag_name | cut -d '"' -f 4)
URL="https://github.com/kutovoys/xray-torrent-blocker/releases/download/${LATEST_RELEASE}/xray-torrent-blocker-${LATEST_RELEASE}-linux-${ARCH}.tar.gz"

curl -sL "$URL" -o tblocker.tar.gz

echo "Extracting files..."
mkdir -p /opt/tblocker
tar -xzf tblocker.tar.gz -C /opt/tblocker --overwrite
rm tblocker.tar.gz

CONFIG_PATH="/opt/tblocker/config.yaml"
CONFIG_TEMPLATE_PATH="/opt/tblocker/config.yaml.example"

if [ ! -f "$CONFIG_PATH" ]; then
  mv "$CONFIG_TEMPLATE_PATH" "$CONFIG_PATH"
  echo "New configuration file created at $CONFIG_PATH"
else
  echo "Configuration file already exists. Checking its contents..."
fi

check_placeholder() {
  local key="$1"
  local value="$2"
  grep -qE "^$key:\s*\"?$value\"?" "$CONFIG_PATH"
}

ask_for_input=true
if (! check_placeholder "AdminBotToken" "ADMIN_" && ! check_placeholder "AdminChatID" "ADMIN_") || check_placeholder "SendAdminMessage" "false"; then
  ask_for_input=false
  echo "Admin bot token and Chat ID are already set in the config. Skipping input."
fi

if $ask_for_input; then
  read -p "Enter the Telegram admin bot token: " admin_bot_token
  read -p "Enter the admin Chat ID: " admin_chat_id

  sed -i "s/ADMIN_BOT_TOKEN/$admin_bot_token/" "$CONFIG_PATH"
  sed -i "s/ADMIN_CHAT_ID/$admin_chat_id/" "$CONFIG_PATH"
fi

echo "Setting up systemd service..."
curl -sL https://raw.githubusercontent.com/kutovoys/xray-torrent-blocker/main/tblocker.service -o /etc/systemd/system/tblocker.service

systemctl daemon-reload
systemctl enable tblocker
systemctl start tblocker

systemctl status tblocker --no-pager

echo ""
echo "==============================================================="
echo ""
echo "Installation complete! The torrent-blocker service is running."
echo "PLEASE SETUP UFW PROPERLY! (https://www.digitalocean.com/community/tutorials/how-to-set-up-a-firewall-with-ufw-on-ubuntu)"
echo "==============================================================="
echo ""
echo "You can configure additional options in the configuration file"
echo "/opt/tblocker/config.yaml"
echo "It is possible to enable sending user notifications via Telegram."
echo ""
echo "==============================================================="