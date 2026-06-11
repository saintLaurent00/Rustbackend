# Build stage
FROM rust:1.81-slim AS builder
WORKDIR /app
COPY . .
RUN cargo build --release

# Run stage
FROM debian:bookworm-slim
RUN apt-get update && apt-get install -y \
    chromium \
    ca-certificates \
    && rm -rf /var/lib/apt/lists/*
WORKDIR /app
COPY --from=builder /app/target/release/render-worker .
ENV CHROME_BIN=/usr/bin/chromium
CMD ["./render-worker"]
