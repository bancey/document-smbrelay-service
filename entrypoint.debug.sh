#!/bin/sh
set -e

# Start sshd as root (required for SSH daemon)
/usr/sbin/sshd

# Run the application as appuser
exec su-exec appuser /app/server
