# Build stage
FROM golang:1.25-alpine AS builder

# Set Go proxy to use alternative mirrors
ENV GOPROXY=https://goproxy.io,https://goproxy.cn,direct
ENV GOSUMDB=off

WORKDIR /app
# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .
# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build \
    -a -installsuffix cgo \
    -ldflags="-s -w" \
    -o main ./cmd/main.go

# Run stage
FROM alpine:latest
# Install ca-certificates for HTTPS
RUN apk --no-cache add ca-certificates
WORKDIR /root/
# Copy the binary from builder
COPY --from=builder /app/main .
# Copy configuration file
COPY --from=builder /app/config.yaml .
# Copy migrations (if needed at runtime)
COPY --from=builder /app/migrations ./migrations
# Expose port
EXPOSE 8080
# Run the application
RUN chmod +x /root/main
CMD ["./main"]