#! /usr/bin/env bash

# Download go modules
go mod download

# Run gpg setup script
bash "$(dirname "$0")/setup-gpg.sh"