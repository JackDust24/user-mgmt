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
	@npx tailwindcss -i ./public/css/input.css -o ./public/css/style.css --watch &
	@echo "Finished Building CSS with Tailwind..."

build: deps css
	@echo "Building Go binary..."
	go build -o bin/$(APP_NAME) ./cmd/server &

# Run the app
run: build
	@echo "Running the server on port $(PORT)..."
	./bin/$(APP_NAME)

init:
	sh scripts/init.sh

templ:
	@echo "Templ generate..."
	@templ generate --watch

air:
	@echo "Starting air..."
	@air

# Uses the rules in .air.toml file to run
server:
	@echo "Starting server and air..."
	@air

airdev:
	@make templ air

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