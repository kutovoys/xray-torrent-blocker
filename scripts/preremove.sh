#!/bin/bash

# XRay Torrent Blocker - Pre-removal script

set -e

echo "XRay Torrent Blocker: Pre-removal script started"

if systemctl is-active --quiet tblocker.service; then
    echo "Stopping tblocker service..."
    systemctl stop tblocker.service
    echo "✓ tblocker service stopped"
else
    echo "tblocker service is not running"
fi

echo "XRay Torrent Blocker: Pre-removal script completed" 