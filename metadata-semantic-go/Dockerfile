# Build stage
FROM golang:1.23-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o metadata-semantic ./cmd/main.go

# Run stage
FROM alpine:latest
WORKDIR /app
COPY --from=builder /app/metadata-semantic .
EXPOSE 8081
CMD ["./metadata-semantic"]
