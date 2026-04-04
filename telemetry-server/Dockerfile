FROM golang:1.25-alpine AS builder

WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY *.go ./
RUN CGO_ENABLED=0 go build -ldflags="-s -w" -o /telemetry-server .

FROM alpine:3.21
RUN apk add --no-cache ca-certificates tzdata
COPY --from=builder /telemetry-server /usr/local/bin/telemetry-server
EXPOSE 9090
ENTRYPOINT ["telemetry-server"]
