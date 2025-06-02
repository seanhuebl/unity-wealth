# .devcontainer/Dockerfile

# ─────────────────────────────────────────────────────────────────────────────
# Stage 1: Start from Microsoft’s Go devcontainer base
# ─────────────────────────────────────────────────────────────────────────────
FROM mcr.microsoft.com/devcontainers/go:1 AS base

# Switch to root to install system packages
USER root

# Install prerequisites, Google Cloud SDK, and Docker CE in one go,
# then clean up apt lists to keep the image lean.
RUN apt-get update && \
    apt-get install -y --no-install-recommends \
    ca-certificates \
    curl \
    gnupg \
    lsb-release \
    apt-transport-https \
    libpq-dev \
    bash-completion \
    postgresql-client && \
    \
    # --- Add Google Cloud SDK apt repository ---
    mkdir -p /usr/share/keyrings && \
    curl -fsSL https://packages.cloud.google.com/apt/doc/apt-key.gpg \
    | gpg --dearmor -o /usr/share/keyrings/cloud.google.gpg && \
    echo "deb [signed-by=/usr/share/keyrings/cloud.google.gpg] \
    https://packages.cloud.google.com/apt cloud-sdk main" \
    > /etc/apt/sources.list.d/google-cloud-sdk.list && \
    \
    # --- Add Docker CE apt repository ---
    curl -fsSL https://download.docker.com/linux/debian/gpg \
    | gpg --dearmor -o /usr/share/keyrings/docker-archive-keyring.gpg && \
    echo "deb [arch=$(dpkg --print-architecture) \
    signed-by=/usr/share/keyrings/docker-archive-keyring.gpg] \
    https://download.docker.com/linux/debian $(lsb_release -cs) stable" \
    > /etc/apt/sources.list.d/docker.list && \
    \
    # Refresh package lists and install SDKs & Docker
    apt-get update && \
    apt-get install -y \
    google-cloud-sdk \
    docker-ce \
    docker-ce-cli \
    containerd.io && \
    \
    # Clean up apt caches
    rm -rf /var/lib/apt/lists/*

# Add "vscode" user to the "docker" group so no sudo is required for `docker ...`
RUN groupadd -f docker && usermod -aG docker vscode

# Copy a repo‑root .bashrc into /home/vscode/.bashrc at build time
# (Assumes you have a .bashrc at the very top‑level of your repo.)
COPY .bashrc /home/vscode/.bashrc
RUN chown vscode:vscode /home/vscode/.bashrc

# Switch back to vscode for Go tool installs and working directory setup
USER vscode
WORKDIR /workspaces/${LOCAL_WORKSPACE_FOLDER_BASENAME}

# Install Go‑based CLIs
RUN go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest \
    && go install github.com/tursodatabase/turso-cli/cmd/turso@latest \
    && go install github.com/seanhuebl/sqlc-qol/v2@v2.0.0 \
    && go install github.com/vektra/mockery/v2@latest \
    && go install golang.org/x/tools/cmd/goimports@latest \
    && go install github.com/pressly/goose/v3/cmd/goose@latest
