# Build stage
FROM golang:1.20-alpine AS builder

# Install git and nodejs (git for go get, nodejs for gopherjs test)
RUN apk add --no-cache git nodejs

# Install GopherJS
RUN go install github.com/gopherjs/gopherjs@v1.20.1

# Set working directory
WORKDIR /app

# Copy all source code first (needed for local packages)
COPY . .

# Tidy modules to ensure local packages are recognized
RUN go mod tidy

# Run tests (allow failures, continue build)
RUN gopherjs test ./game/...

# Build with GopherJS (conditionally enable debug logging)
RUN gopherjs build -o game.js .

# Build the server (with embedded index.html)
RUN CGO_ENABLED=0 GOOS=linux go build -o starship-server ./server

# Production stage - minimal runtime with Go server
FROM alpine:latest

# Add ca-certificates for HTTPS if needed
RUN apk --no-cache add ca-certificates

WORKDIR /app

# Copy built game files (game.js needed at runtime)
COPY --from=builder /app/game.js .
COPY --from=builder /app/game.js.map .

# Copy the server binary (index.html is embedded)
COPY --from=builder /app/starship-server .

# Expose HTTP port and TURN port (UDP/TCP)
EXPOSE 8080
EXPOSE 3478/udp
EXPOSE 3478/tcp

# Run the server
CMD ["./starship-server", "-port", "8080", "--turn-port", "3478", "-static", "/app"]