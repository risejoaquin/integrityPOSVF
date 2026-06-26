#!/bin/bash
set -e

echo "Starting installation of integrityPOS..."

UPDATE_ONLY=0
if [ "$1" == "--update" ]; then
    UPDATE_ONLY=1
    echo "Update mode enabled."
fi

# Detect OS / binary extension
BIN_NAME="posd"
if [ -f "posd.exe" ]; then
    BIN_NAME="posd.exe"
fi

if [ ! -f "$BIN_NAME" ]; then
    echo "Error: $BIN_NAME not found."
    exit 1
fi

chmod +x $BIN_NAME

INSTALL_DIR="/opt/integritypos"
if [[ "$OSTYPE" == "msys" || "$OSTYPE" == "cygwin" ]]; then
    INSTALL_DIR="C:/integritypos"
fi

if [ -d "$INSTALL_DIR" ]; then
    if [ $UPDATE_ONLY -eq 0 ]; then
        echo "Directory $INSTALL_DIR already exists."
        read -p "Do you want to update the existing installation? (y/n) " -n 1 -r
        echo
        if [[ ! $REPLY =~ ^[Yy]$ ]]; then
            echo "Installation aborted."
            exit 0
        fi
    fi
fi

mkdir -p "$INSTALL_DIR"

if [ $UPDATE_ONLY -eq 1 ]; then
    cp "$BIN_NAME" "$INSTALL_DIR/"
    cp -r migrations "$INSTALL_DIR/" 2>/dev/null || true
    echo "Updated binary and migrations successfully."
else
    cp -r * "$INSTALL_DIR/"
    echo "Installed successfully to $INSTALL_DIR"
fi

echo "To run: cd $INSTALL_DIR && ./$BIN_NAME"
