FROM golang:latest as builder

WORKDIR /usr/src/app

# Show Go toolchain info
RUN set -eux; \
    echo "[builder] Go version:"; go version; \
    echo "[builder] Go env:"; go env

# Copy module files first for better caching
COPY go.mod go.sum ./
RUN go mod download && go mod verify
COPY . .
RUN go build -v -o /run-app .


FROM debian:bookworm

# Install CA bundle so HTTPS (Supabase) certs verify correctly
RUN set -eux; \
    apt-get update; \
    apt-get install -y --no-install-recommends ca-certificates; \
    update-ca-certificates; \
    rm -rf /var/lib/apt/lists/*

# Ensure Go uses system cert bundle path
ENV SSL_CERT_FILE=/etc/ssl/certs/ca-certificates.crt

COPY --from=builder /run-app /usr/local/bin/
CMD ["run-app"]
