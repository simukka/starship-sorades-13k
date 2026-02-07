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

# Run tests
RUN gopherjs test ./game/...

# Build with GopherJS (conditionally enable debug logging)
RUN gopherjs build -o game.js .

# Production stage - serve static files
FROM nginx:alpine

# Copy built game and static files
COPY --from=builder /app/game.js /usr/share/nginx/html/
COPY --from=builder /app/game.js.map /usr/share/nginx/html/
COPY index.html /usr/share/nginx/html/
COPY jsfxr.js /usr/share/nginx/html/
COPY starship-sorades.css /usr/share/nginx/html/

# Expose port 80
EXPOSE 80

CMD ["nginx", "-g", "daemon off;"]