# Xray Torrent Blocker

[![en](https://img.shields.io/badge/lang-en-red)](https://github.com/kutovoys/xray-torrent-blocker/blob/main/README.md)
[![ru](https://img.shields.io/badge/lang-ru-blue)](https://github.com/kutovoys/xray-torrent-blocker/blob/main/README.ru.md)

Xray Torrent Blocker is an application designed to block torrent usage by users of the [Xray-based](https://github.com/XTLS/Xray-core) panels. The application analyzes logs, detects torrent activity, and temporarily blocks the user, sending webhooks to the configured webhook URL.

## Features

- Monitoring logs of nodes and the panel for torrent usage
- IP address blocking at the system level with maximum block speed (no abuse reports!)
- Connection termination via conntrack - instantly break existing torrent connections
- Sending webhooks to the configured webhook URL
- Configurable through a configuration file
- Supports various firewalls for blocking (iptables, nftables)
- Configurable block duration
- Supports temporary blocking with automatic unblocking
- Install with apt or yum package managers
- Persistent block state between application restarts
- Automatic block restoration after system reboot
- Automatic cleanup of expired blocks

## Requirements

- Firewall (iptables or nftables)
- Xray log file with enabled logging

## Installation

### Quick Install Script

The easiest way to install Xray Torrent Blocker is using the installation script:

```bash
bash <(curl -fsSL git.new/install)
```

This script will automatically:

- Detect your system architecture
- Download the latest release
- Install the binary to `/opt/tblocker/`
- Create a default configuration file
- Set up the systemd service
- Start the service

During installation, you will be prompted to enter the path to your log file and select your preferred firewall (iptables, or nftables). Other configuration parameters can be adjusted manually by editing `/opt/tblocker/config.yaml` if needed.

### From Package Repository

After installation from the repository, a default configuration will be created at `/opt/tblocker/config.yaml`.

For basic operation, you only need to change `LogFile` to point to your xray logs path.

A systemd service `tblocker.service` will also be created for automatic startup at system boot. Automatic startup will be enabled. You just need to start the service after editing the config:

```bash
systemctl start tblocker
```

#### Debian/Ubuntu Based Systems

```bash
apt update && apt install -y curl gnupg
curl https://repo.remna.dev/xray-tools/public.gpg | gpg --yes --dearmor -o /usr/share/keyrings/openrepo-xray-tools.gpg
echo "deb [arch=any signed-by=/usr/share/keyrings/openrepo-xray-tools.gpg] https://repo.remna.dev/xray-tools/ stable main" > /etc/apt/sources.list.d/openrepo-xray-tools.list
apt update
apt install tblocker
```

#### RPM Based Systems

```bash
echo """
[xray-tools-rpm]
name=xray-tools-rpm
baseurl=https://repo.remna.dev/xray-tools-rpm
enabled=1
repo_gpgcheck=1
gpgkey=https://repo.remna.dev/xray-tools-rpm/public.gpg
""" > /etc/yum.repos.d/xray-tools-rpm.repo
yum update
yum install tblocker
```

### From Releases Binary

1. Install required dependencies:
   ```bash
   # For Debian/Ubuntu
   sudo apt install conntrack
   # For CentOS/RHEL
   sudo yum install conntrack-tools
   ```
2. Download the latest release for your architecture from [GitHub Releases](https://github.com/kutovoys/xray-torrent-blocker/releases)
3. Extract the binary and make it executable:
   ```bash
   chmod +x tblocker
   ```
4. Move to system directory:
   ```bash
   sudo mv tblocker /opt/tblocker/
   ```
5. Create config file `/opt/tblocker/config.yaml` with your settings
6. Copy [systemd service file](tblocker.service) to `/etc/systemd/system/tblocker.service` and start the service
   ```bash
   sudo systemctl daemon-reload
   sudo systemctl enable tblocker
   sudo systemctl start tblocker
   ```

## Configuration

### Default Configuration

After installation, the application uses default configuration. You can customize it by editing `/opt/tblocker/config.yaml`:

```yaml
# Log file to monitor
LogFile: "/var/log/remnanode/access.log"

# Block duration in minutes
BlockDuration: 10

# Tag used to identify torrent traffic in logs
TorrentTag: "TORRENT"

# Firewall to use for blocking (iptables, nft)
BlockMode: "iptables"
```

### Advanced Configuration

For advanced usage, you can configure additional features:

```yaml
# IP addresses to bypass blocking
BypassIPS:
  - "127.0.0.1"
  - "::1"

# Storage directory for block data
StorageDir: "/opt/tblocker"

# Username processing regex for webhooks
UsernameRegex: "^(.+)$"

# Webhook configuration
SendWebhook: false
WebhookURL: "https://your-webhook-url.com/endpoint"
WebhookTemplate: '{"username":"%s","ip":"%s","server":"%s","action":"%s","duration":%d,"timestamp":"%s"}'
WebhookHeaders:
  Authorization: "Bearer your-token"
  Content-Type: "application/json"
```

## Panels Configuration

### For all panels

1. Configure bittorrent traffic tagging. Section `routing`. Add the rule:

   ```json
   {
     "protocol": ["bittorrent"],
     "outboundTag": "TORRENT",
     "type": "field"
   }
   ```

   Here, `TORRENT` is the tag that the application will use to filter logs.

2. Configure bittorrent traffic blocking. Section `outbounds`. Send all traffic to blackhole:

   ```json
   {
     "protocol": "blackhole",
     "tag": "TORRENT"
   }
   ```

### Remnawave

1. Create the log directory:

   ```bash
   mkdir -p /var/log/remnanode
   ```

2. Add volume to remnanode's `docker-compose.yml`:

   ```yaml
   volumes:
     - "/var/log/remnanode:/var/log/remnanode"
   ```

3. Setup logging in xray config:

   ```json
   "log": {
       "error": "/var/log/remnanode/error.log",
       "access": "/var/log/remnanode/access.log",
       "loglevel": "error"
   }
   ```

4. Restart the remnanode.

### Marzban

1. Create the log directory:

   ```bash
   mkdir -p /var/lib/marzban-node
   ```

2. Add volume to marzban-node's `docker-compose.yml`:

   ```yaml
   volumes:
     - /var/lib/marzban-node:/var/lib/marzban-node
   ```

3. Setup logging in xray config:

   ```json
   "log": {
       "error": "/var/lib/marzban-node/error.log",
       "access": "/var/lib/marzban-node/access.log",
       "loglevel": "error"
   }
   ```

4. Set UsernameRegex value in config.yaml:

   ```yaml
   UsernameRegex: "^\\d+\\.(.+)$"
   ```

5. Restart the marzban-node.

### Other Panels

For other Xray-based panels, ensure that:

1. Log files are accessible on the host system
2. Log format includes necessary information (IP, user identification)
3. Bittorrent traffic is properly tagged in routing rules

## Tips

### Working Behind a TCP Proxy

⚠️ **Important**: If you place Nginx/HAProxy/another TCP proxy in front of Xray, make sure the real client IP reaches Xray via the PROXY protocol; otherwise, you may end up blocking 127.0.0.1 or your server IP instead of the actual offender.

**Xray Configuration Example:**

```json
{
  "inbounds": [
    {
      "port": 444,
      "protocol": "vless",
      "streamSettings": {
        "network": "tcp",
        "security": "reality",
        "sockopt": {
          "acceptProxyProtocol": true // accept PROXY v1/v2 from the proxy
        }
      }
    }
  ]
}
```

**Nginx Configuration Example:**

```nginx
stream {
    server {
        listen 443;
        proxy_pass 127.0.0.1:444;  # your Xray inbound
        proxy_protocol on;         # send PROXY protocol to backend
    }
}
```

**HAProxy Configuration Example:**

```
backend xray_backend
    mode tcp
    server xray1 127.0.0.1:444 send-proxy-v2
```

This ensures that Xray receives the real client IP address in its access logs, allowing tblocker to block the correct IP addresses.

### Reading logs

To read `tblocker` logs, you can use the following command:

```bash
journalctl -u tblocker -f --no-pager
```

### Logrotate Configuration

To prevent log files from consuming too much disk space, configure logrotate:

```bash
sudo bash -c 'cat > /etc/logrotate.d/remnanode <<EOF
/var/log/remnanode/*.log {
    size 50M
    rotate 5
    compress
    missingok
    notifempty
    copytruncate
}
EOF'
```

### Working with Webhooks

Webhooks allow you to integrate tblocker with external systems:

- **Panel**: Enable/Disable user in Panel for blocking on all nodes
- **Telegram**: Send notifications to Telegram groups. For admin and user notifications.
- **Custom APIs**: Connect to your own monitoring systems

For receiving webhooks, you can use [n8n](https://n8n.io/) or any other webhook service.

## Contributing

We welcome contributions from the community! If you have ideas for improvements or have found a bug, please:

1. Create an issue on GitHub
2. Fork the repository
3. Create a feature branch
4. Make your changes
5. Submit a pull request

For major changes, please open an issue first to discuss what you would like to change.

## VPN Recommendation

For secure and reliable internet access, we recommend [BlancVPN](https://getblancvpn.com/?ref=tblocker). Use promo code `TRYBLANCVPN` for 15% off your subscription.
