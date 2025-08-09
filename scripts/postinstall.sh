#!/bin/bash

# XRay Torrent Blocker - Post-installation script

set -e

echo "XRay Torrent Blocker: Post-installation script started"

if [ ! -d "/opt/tblocker" ]; then
    mkdir -p /opt/tblocker
    chmod 755 /opt/tblocker
fi

chmod 755 /opt/tblocker/tblocker
chmod 644 /opt/tblocker/config.yaml

systemctl daemon-reload

systemctl enable tblocker.service
echo "✓ XRay Torrent Blocker service installed and enabled to start on boot"

echo "
╔════════════════════════════════════════════════════════════════════════════╗
║                          !!! IMPORTANT NOTICE !!!                          ║
║                                                                            ║
║  Before starting the service, you MUST:                                    ║
║                                                                            ║
║  1. Configure the correct log file path in /opt/tblocker/config.yaml       ║
║     Current default path may not work for your setup!                      ║
║                                                                            ║
║  2. Review the documentation at:                                           ║
║     https://github.com/kutovoys/xray-torrent-blocker                       ║
║                                                                            ║
║  3. Configure additional parameters according to your needs                ║
║                                                                            ║
║  After configuration is complete, start the service with:                  ║
║  systemctl start tblocker.service                                          ║
║                                                                            ║
╚════════════════════════════════════════════════════════════════════════════╝
"