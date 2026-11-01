#!/usr/bin/env bash
# Grant the current user read access to the BPF capture devices so the GUI app
# (which must NOT run as root) can capture live traffic. This is a development
# convenience: it lasts until the next reboot. The permanent fix is the
# ChmodBPF-style LaunchDaemon helper tracked in TASK-9.
#
# Run with sudo:  sudo scripts/grant-capture.sh   (or: make grant-capture)
set -euo pipefail

if [[ "$(uname)" != "Darwin" ]]; then
  echo "grant-capture: macOS only. On Linux use: sudo setcap cap_net_raw,cap_net_admin+eip <binary>" >&2
  exit 1
fi
if [[ "$(id -u)" -ne 0 ]]; then
  echo "grant-capture: must run as root — re-run with: sudo scripts/grant-capture.sh" >&2
  exit 1
fi

user_name="${SUDO_USER:-$USER}"
count=0
for dev in /dev/bpf*; do
  [[ -e "$dev" ]] || continue
  chown "$user_name" "$dev" 2>/dev/null && count=$((count + 1)) || true
done

echo "grant-capture: handed $count /dev/bpf* device(s) to '$user_name' (until next reboot)."
echo "Now run:  make run    then click Start capture in the app."
