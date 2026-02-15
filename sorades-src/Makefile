# Makefile for Starship Sorades GopherJS build

.PHONY: all build build-debug clean

# Default target
all: build

# Build the game.js file using GopherJS
build: 
	docker build -t starship-sorades .

# Build with debug logging enabled
build-debug:
	docker build --build-arg ENABLE_DEBUG=true -t starship-sorades:debug .

# Clean build artifacts
clean:
	rm -f game.js game.js.map

# Run with host networking (recommended for local development)
# This allows TURN server to use the actual host IP
serve: build
	docker run --network host starship-sorades:latest

# Run with port mapping and explicit IP (for production/cloud deployment)
# Usage: make serve-ip PUBLIC_IP=10.0.1.195
PUBLIC_IP ?= 127.0.0.1
serve-ip: build
	docker run -p 8080:8080 -p 3478:3478/udp -p 3478:3478/tcp starship-sorades:latest \
		./starship-server -port 8080 --turn-port 3478 --public-ip $(PUBLIC_IP) -static /app