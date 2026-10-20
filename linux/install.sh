#!/usr/bin/env bash
# Sloth Mode Linux installer — no root privileges, no account needed.
set -euo pipefail
HERE="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" && pwd)"
case "$(uname -s)" in Linux) ;; *) echo 'This package is for Linux.' >&2; exit 1;; esac
case "$(uname -m)" in x86_64) machine=amd64;; aarch64) machine=arm64;; *) echo 'Only Linux x86_64 and aarch64 are packaged.' >&2; exit 1;; esac
SRC="$HERE/bin/sloth-mode-connector-linux-$machine"
[[ -f "$SRC" ]] || { echo "Missing connector executable: $SRC" >&2; exit 1; }
DATA_BASE="${XDG_DATA_HOME:-$HOME/.local/share}"
CONFIG_BASE="${XDG_CONFIG_HOME:-$HOME/.config}"
INSTALL="$DATA_BASE/sloth-mode/bin"
UNIT_DIR="$CONFIG_BASE/systemd/user"
APP_DIR="$DATA_BASE/applications"
ICON_DIR="$DATA_BASE/icons/hicolor/scalable/apps"
mkdir -p "$INSTALL" "$UNIT_DIR" "$APP_DIR" "$ICON_DIR"
DEST="$INSTALL/sloth-mode-connector"
# Stop the previous service before replacing the binary.
if command -v systemctl >/dev/null 2>&1; then systemctl --user stop sloth-mode.service 2>/dev/null || true; fi
install -m 0755 "$SRC" "$DEST"
install -m 0644 "$HERE/sloth-mode.svg" "$ICON_DIR/sloth-mode.svg"
printf '[Unit]\nDescription=Sloth Mode local AI connector (account optional)\nAfter=network.target\n\n[Service]\nType=simple\nExecStart="%s"\nRestart=on-failure\nRestartSec=5\nNoNewPrivileges=true\nUMask=0077\n\n[Install]\nWantedBy=default.target\n' "$DEST" > "$UNIT_DIR/sloth-mode.service"
cat > "$APP_DIR/sloth-mode.desktop" <<DESKTOP
[Desktop Entry]
Type=Application
Name=Sloth Mode — Setup & Pairing
Comment=Local AI connector, no account required
Exec=$DEST --setup
Icon=sloth-mode
Terminal=true
Categories=Utility;Office;
Keywords=AI;Gemma;Productivity;
StartupNotify=false
DESKTOP
chmod 0644 "$APP_DIR/sloth-mode.desktop"
if command -v systemctl >/dev/null 2>&1 && systemctl --user daemon-reload 2>/dev/null; then
  systemctl --user enable --now sloth-mode.service || { echo 'Could not start user service. You can run the connector manually.' >&2; }
else
  echo 'No systemd user session found; starting connector for this login.'
  (nohup "$DEST" >/dev/null 2>&1 </dev/null &)
fi
printf '\nSloth Mode installed for your user.\nNo cloud account is required.\nOpen https://sloth.marceli.dev/dashboard and pair this computer once.\n\n'
echo 'For a visible setup screen: open “Sloth Mode — Setup & Pairing” from your app menu.'
echo 'Or run: '"$DEST"' --setup'
echo 'Hardware report: '"$DEST"' --hardware'
echo 'Optional verified Gemma download: '"$DEST"' --install-ai'
echo 'Optional cloud account linking: '"$DEST"' --link-account CODE'
if [[ "${1:-}" != "--no-setup" ]] && [[ -n "${DISPLAY:-}${WAYLAND_DISPLAY:-}" ]]; then "$DEST" --setup || true; fi
