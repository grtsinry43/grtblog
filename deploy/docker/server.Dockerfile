FROM golang:1.24-alpine AS builder

WORKDIR /src/server

RUN apk add --no-cache ca-certificates git

COPY server/go.mod server/go.sum ./
RUN go mod download

COPY server/. .

RUN GOBIN=/out go install github.com/pressly/goose/v3/cmd/goose@latest

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
  go build -trimpath -ldflags="-s -w" -o /out/grtblog-server ./cmd/api

FROM alpine:3.21 AS runtime

RUN apk add --no-cache ca-certificates tzdata su-exec \
  && addgroup -g 10001 -S app \
  && adduser -u 10001 -S app -G app

WORKDIR /app

COPY --from=builder /out/grtblog-server /app/grtblog-server
COPY --from=builder /out/goose /usr/local/bin/goose
COPY --from=builder /src/server/docs /app/docs
COPY --from=builder /src/server/migrations /app/migrations
COPY deploy/docker/server-entrypoint.sh /usr/local/bin/server-entrypoint.sh

RUN mkdir -p /app/storage/html /app/storage/uploads /app/storage/geoip \
  && chown -R app:app /app \
  && chmod +x /usr/local/bin/server-entrypoint.sh

EXPOSE 8080

ENTRYPOINT ["/usr/local/bin/server-entrypoint.sh"]
