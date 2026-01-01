FROM docker.io/library/golang:1.25.4 AS builder

# パッケージダウンロード
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

# ビルド
COPY . .
RUN CGO_ENABLED=0 go build -o app ./cmd/main.go

FROM docker.io/library/alpine:3.23.2

# Atlasをインストール
COPY --from=arigaio/atlas:1.0.0 /atlas /usr/local/bin/atlas

# 実行環境構築
WORKDIR /app
COPY --from=builder /app/app .
COPY migrations ./migrations
COPY docker/entrypoint.sh /entrypoint.sh

# Tailscaleをインストール
RUN apk update \
	&& apk add --no-cache tailscale=1.90.9-r1 iptables=1.8.11-r1 \
	&& chmod +x /entrypoint.sh

ENTRYPOINT ["/entrypoint.sh"]
CMD ["/app/app"]
