# Variables
APP_NAME=user-mgmt
PORT=8080

# Default target
all: build run

# Install dependencies
deps:
	@echo "Installing Go dependencies..."
	# go mod tidy

css:
	@echo "Building CSS with Tailwind..."
	npx tailwindcss -o ./public/css/style.css --minify

build: deps css
	@echo "Building Go binary..."
	go build -o bin/$(APP_NAME) ./cmd/server

# Run the app
run: build
	@echo "Running the server on port $(PORT)..."
	./bin/$(APP_NAME)

init:
	sh scripts/init.sh

templ:
	@templ generate --watch

# Uses the rules in .air.toml file to run
server:
	@air

dev:
	@make -j3 css templ server

# Clean up build artifacts
clean:
	@echo "Cleaning up..."
	rm -rf bin/

# Watch for changes
watch:
	@echo "Watching for changes..."
	@while true; do \
		inotifywait -e modify,create,delete -r ./; \
		make run; \
	done