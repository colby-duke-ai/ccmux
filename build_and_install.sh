#!/bin/bash
set -e

cd "$(dirname "$0")"
go build -o ccmux ./cmd/ccmux
rm -f ~/bin/ccmux
mv ccmux ~/bin/ccmux
echo "Installed ccmux to ~/bin/ccmux"

for session in $(tmux list-sessions -F '#S' 2>/dev/null | grep '^ccmux-'); do
	tmux kill-session -t "$session"
	echo "Killed session: $session"
done
