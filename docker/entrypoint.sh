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

# proxychainsの設定ファイルを作成
cat <<EOT > /etc/proxychains.conf
strict_chain
proxy_dns
remote_dns_resolv_conf
[ProxyList]
socks5 127.0.0.1 1055
EOT

# アプリケーション起動
exec proxychains4 -f /etc/proxychains.conf "$@"
