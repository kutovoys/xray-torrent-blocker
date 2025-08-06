#!/bin/bash

# XRay Torrent Blocker - Post-removal script

set -e

echo "XRay Torrent Blocker: Post-removal script started"

if systemctl is-enabled --quiet tblocker.service; then
    echo "Disabling tblocker service..."
    systemctl disable tblocker.service
    echo "✓ tblocker service disabled"
else
    echo "tblocker service is not enabled"
fi

systemctl daemon-reload

echo "Xray Torrent Blocker: Post-removal script completed" 