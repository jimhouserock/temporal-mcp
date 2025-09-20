# Use Go 1.24 as base image
FROM golang:1.24-alpine AS builder

# Set working directory
WORKDIR /app

# Install build dependencies
RUN apk add --no-cache git make

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies and update go.sum for HTTP transport
RUN go mod download
RUN go get github.com/metoro-io/mcp-golang/transport/http@v0.11.0
RUN go mod tidy

# Copy source code
COPY . .

# Build the application
RUN make build

# Use minimal alpine image for runtime
FROM alpine:latest

# Install ca-certificates for HTTPS requests
RUN apk --no-cache add ca-certificates

# Set working directory
WORKDIR /root/

# Copy the binary from builder stage
COPY --from=builder /app/bin/temporal-mcp .

# Copy config sample (optional, for reference)
COPY --from=builder /app/config.sample.yml .

# Expose port (Smithery will set PORT environment variable to 8081)
EXPOSE 8081

# Set environment variable for port
ENV PORT=8081

# Start the server
# Note: You may need to modify your Go application to:
# 1. Listen on HTTP instead of stdio
# 2. Implement /mcp endpoint
# 3. Handle CORS headers
# 4. Listen on PORT environment variable
CMD ["./temporal-mcp", "--config", "config.sample.yml"]
