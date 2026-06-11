# Build stage
FROM golang:1.23-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o gateway-auth ./cmd/main.go

# Run stage
FROM alpine:latest
WORKDIR /app
COPY --from=builder /app/gateway-auth .
EXPOSE 8080
CMD ["./gateway-auth"]
