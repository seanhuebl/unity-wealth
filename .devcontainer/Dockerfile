# .devcontainer/Dockerfile

# ─────────────────────────────────────────────────────────────────────────────
# Stage 1: Start from Microsoft’s Go devcontainer base
# ─────────────────────────────────────────────────────────────────────────────
FROM mcr.microsoft.com/devcontainers/go:1 AS base

# Install packages needed by gcloud + any CGO drivers
USER root
RUN apt-get update && \
    apt-get install -y \
    curl \
    apt-transport-https \
    ca-certificates \
    gnupg \
    libpq-dev \
    postgresql-client \
    && rm -rf /var/lib/apt/lists/*

# Install Google Cloud SDK
RUN curl https://packages.cloud.google.com/apt/doc/apt-key.gpg \
    | apt-key add - \
    && echo "deb [signed-by=/usr/share/keyrings/cloud.google.gpg] \
    http://packages.cloud.google.com/apt cloud-sdk main" \
    > /etc/apt/sources.list.d/google-cloud-sdk.list \
    && apt-get update \
    && apt-get install -y google-cloud-sdk \
    && rm -rf /var/lib/apt/lists/*

# Install Go‑based CLIs (sqlc, turso, sqlc-qol, mockery, goimports, goose)
USER vscode
RUN go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest \
    && go install github.com/turso/cli/cmd/turso@latest \
    && go install github.com/SeanHuebl/sqlc-qol/v2@v2.0.0 \
    && go install github.com/vektra/mockery/v2@latest \
    && go install golang.org/x/tools/cmd/goimports@latest \
    && go install github.com/pressly/goose/v3/cmd/goose@latest

# Copy a repository‑root .bashrc into /home/vscode/.bashrc at build time
# (Assumes you have a .bashrc at the very top‐level of your repo.)
COPY .bashrc /home/vscode/.bashrc
RUN chown vscode:vscode /home/vscode/.bashrc

# Set working directory (inside the container) to the mounted workspace
WORKDIR /workspaces/${LOCAL_WORKSPACE_FOLDER_BASENAME}
