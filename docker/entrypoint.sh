#!/bin/sh
set -e

# Tailscale起動
tailscaled --tun=userspace-networking &

# Tailscale接続
# HEADSCALE_URLが設定されていない場合はデフォルトのTailscaleサーバーに接続
if [ -z "$HEADSCALE_URL" ]; then
  tailscale up --authkey="${TAILSCALE_AUTHKEY}"
else
  tailscale up --authkey="${TAILSCALE_AUTHKEY}" --login-server="${HEADSCALE_URL}"
fi

# アプリケーション起動
exec "$@"
