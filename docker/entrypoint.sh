#!/bin/sh
set -e

# Tailscale起動
tailscaled --tun=userspace-networking --socks5-server=localhost:1055 --outbound-http-proxy-listen=localhost:1055 &

# Tailscale接続
# HEADSCALE_URLが設定されていない場合はデフォルトのTailscaleサーバーに接続
if [ -z "$HEADSCALE_URL" ]; then
  tailscale up --authkey="${TAILSCALE_AUTHKEY}"
else
  tailscale up --authkey="${TAILSCALE_AUTHKEY}" --login-server="${HEADSCALE_URL}" --accept-routes
fi

# Atlas DBマイグレーション実行
DB_URL="maria://${NS_MARIADB_USER}:${NS_MARIADB_PASSWORD}@${NS_MARIADB_HOSTNAME}:${NS_MARIADB_PORT}/${NS_MARIADB_DATABASE}"
atlas migrate apply \
  --url "${DB_URL}" \
  --dir "file:///app/migrations"

# プロキシ設定
export HTTP_PROXY=http://localhost:1055/
export http_proxy=http://localhost:1055/
export HTTPS_PROXY=http://localhost:1055/
export https_proxy=http://localhost:1055/

# アプリケーション起動
exec "$@"
