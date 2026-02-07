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

serve: build
	docker run -p 8080:80 starship-sorades:latest
# watch: install-gopherjs
# 	@echo "Watching for changes..."
# 	@while true; do \
# 		find . -name '*.go' | entr -d make build; \
# 	done