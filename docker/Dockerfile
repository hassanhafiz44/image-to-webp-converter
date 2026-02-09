# docker/Dockerfile

# --- Build stage ---
FROM rust:1.85-bookworm AS builder

WORKDIR /build

# Cache dependencies: copy manifests, create dummy source, build deps
COPY Cargo.toml Cargo.lock* ./
RUN mkdir src && echo "fn main() {}" > src/main.rs
RUN cargo build --release 2>/dev/null || true
RUN rm -rf src

# Copy real source and build
COPY src/ ./src/
RUN touch src/main.rs && cargo build --release

# --- Runtime stage ---
FROM debian:bookworm-slim

LABEL maintainer="Avas" \
      description="Rust Image to WebP Converter"

RUN apt-get update && apt-get install -y --no-install-recommends \
    ca-certificates tzdata \
    && rm -rf /var/lib/apt/lists/* \
    && groupadd --system app && useradd --system --gid app app

WORKDIR /app

# Copy the compiled binary from builder
COPY --from=builder /build/target/release/image-converter /app/converter

RUN mkdir -p /app/images /app/output \
    && chown -R app:app /app

USER app

# Default command runs the converter
CMD ["/app/converter"]
