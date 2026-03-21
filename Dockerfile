# ── Build stage ──────────────────────────────────────────
FROM golang:alpine AS builder

WORKDIR /build

# tzdata と ca-certificates をインストール（scratch へのコピー用）
RUN apk add --no-cache tzdata ca-certificates

# 依存関係を先にキャッシュ
COPY go.mod go.sum ./
RUN go mod download

COPY . .

# 静的リンクバイナリを生成（scratch で動作させるため CGO 無効）
# -s -w でデバッグ情報・シンボルテーブルを除去してバイナリを最小化
RUN CGO_ENABLED=0 GOOS=linux go build \
    -ldflags="-s -w" \
    -trimpath \
    -o server .

# ── Final stage ──────────────────────────────────────────
# scratch: OS レイヤーなし、最小サイズ・最速起動
FROM scratch

# タイムゾーン情報と CA 証明書をビルドステージからコピー
COPY --from=builder /usr/share/zoneinfo /usr/share/zoneinfo
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

COPY --from=builder /build/server /app/server

EXPOSE 8080

WORKDIR /app

ENTRYPOINT ["/app/server"]
