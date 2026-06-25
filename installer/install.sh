#!/bin/bash
set -e

echo "Starting installation of integrityPOS..."

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

mkdir -p "$INSTALL_DIR"
cp -r * "$INSTALL_DIR/"

echo "Installed successfully to $INSTALL_DIR"
echo "To run: cd $INSTALL_DIR && ./$BIN_NAME -config config.yaml"
