#!/bin/bash

# Function to shut down Redis when the script exits
cleanup() {
    echo "Shutting down Redis..."
    # Shut down Redis gracefully via redis-cli
    redis-cli shutdown
}

# Trap exit signals to run cleanup()
trap cleanup EXIT SIGINT SIGTERM

# Start Redis in daemon mode
redis-server --daemonize yes

# Optionally, wait for Redis to be ready
sleep 2

# Run the Go server; this will block until the server stops.
go run .

# When go run finishes, the cleanup() function will be triggered by the trap.
