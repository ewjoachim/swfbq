# Build stage
FROM golang:1.21 AS builder

WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . ./

# Build the binary
RUN CGO_ENABLED=0 GOOS=linux go build -o swfbq

# Final stage
FROM ubuntu:22.04

# Install CA certificates and set noninteractive mode
ENV DEBIAN_FRONTEND=noninteractive
RUN apt-get update && \
    apt-get install -y ca-certificates && \
    rm -rf /var/lib/apt/lists/*

WORKDIR /app

# Copy the binary from builder
COPY --from=builder /app/swfbq .

# Set the entrypoint
ENTRYPOINT ["/app/swfbq"]
