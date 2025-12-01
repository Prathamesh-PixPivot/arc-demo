# 🔨 Build Stage
FROM golang:1.24 AS builder

WORKDIR /app

# Cache dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# 🔑 Generate Dummy Keys (for Dev/Demo mode)
# We generate them here so they are baked into the image
RUN openssl genrsa -out private.pem 2048 && \
    openssl rsa -in private.pem -outform PEM -pubout -out public.pem

# Build all binaries
RUN mkdir -p bin && \
    for cmd in server genkey migrate retry setup-tenant consentctl; do \
    CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o bin/$cmd ./cmd/$cmd; \
    done

# 🏃 Runtime Stage
# Using debian-slim instead of distroless for better debugging tools if needed, 
# but distroless is fine if we have everything. Let's stick to distroless for security.
FROM gcr.io/distroless/base-debian11

WORKDIR /app

# Copy Binaries
COPY --from=builder /app/bin /app/bin
# Copy Config (if any static config files exist)
COPY --from=builder /app/config /app/config
# Copy Generated Keys
COPY --from=builder /app/private.pem /app/private.pem
COPY --from=builder /app/public.pem /app/public.pem
# Copy Documentation (Fixes the panic)
COPY docs /app/docs

# Set the default command to run the API Server
CMD ["/app/bin/server"]
