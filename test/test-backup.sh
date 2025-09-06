#!/bin/sh
set -e
echo "[test-backup] Running CLI to generate file..."
/usr/local/bin/composectl gen-backup-meta -i /compose.backup.yml -o /backup
echo "[test-backup] Done."
