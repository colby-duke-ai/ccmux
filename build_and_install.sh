#!/bin/bash
set -e

cd "$(dirname "$0")"
go build -o ccmux ./cmd/ccmux
rm -f ~/bin/ccmux
mv ccmux ~/bin/ccmux
echo "Installed ccmux to ~/bin/ccmux"
