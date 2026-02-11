# docker/Dockerfile

# --- Build stage ---
FROM golang:1.23-bookworm AS builder

RUN apt-get update && apt-get install -y --no-install-recommends \
    libwebp-dev gcc libc6-dev \
    && rm -rf /var/lib/apt/lists/*

WORKDIR /build

# Copy source and resolve dependencies
COPY go.mod ./
COPY main.go ./
RUN go mod tidy && go mod download

# Build
RUN CGO_ENABLED=1 go build -ldflags="-s -w" -o converter .

# --- Runtime stage ---
FROM debian:bookworm-slim

LABEL maintainer="Avas" \
      description="Go Image to WebP Converter"

RUN apt-get update && apt-get install -y --no-install-recommends \
    ca-certificates tzdata libwebp7 \
    && rm -rf /var/lib/apt/lists/* \
    && groupadd --system app && useradd --system --gid app app

WORKDIR /app

# Copy the compiled binary from builder
COPY --from=builder /build/converter /app/converter

RUN mkdir -p /app/images /app/output \
    && chown -R app:app /app

USER app

# Default command runs the converter
CMD ["/app/converter"]
