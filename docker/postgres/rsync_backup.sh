#!/bin/bash

# Configuration
SOURCE_USER=prod_user
SOURCE_HOST=prod.example.com
SOURCE_DIR=/home/user/backups/
DEST_DIR=/home/user/backups/
SSH_PORT=22  # change if custom port

# Run rsync over SSH
rsync -avz -e "ssh -p $SSH_PORT" "${SOURCE_USER}@${SOURCE_HOST}:${SOURCE_DIR}" "$DEST_DIR"
