# Build stage
FROM rust:1.81-slim AS builder
WORKDIR /app
COPY . .
RUN cargo build --release

# Run stage
FROM debian:bookworm-slim
WORKDIR /app
COPY --from=builder /app/target/release/orchestrator-sync-rust .
CMD ["./orchestrator-sync-rust"]
