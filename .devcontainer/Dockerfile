# .devcontainer/Dockerfile
FROM mcr.microsoft.com/devcontainers/go:1 AS base

# Capture the build‑arg (will be non‑empty in Codespaces, empty locally)
ARG GPG_PRIVATE_KEY_ASC

USER root
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

# If GPG_PRIVATE_KEY_ASC was provided (Codespaces), import it now.
# If it’s empty (local build), skip.
RUN if [ -n "${GPG_PRIVATE_KEY_ASC}" ]; then \
      mkdir -p /home/vscode/.gnupg && \
      chown vscode:vscode /home/vscode/.gnupg && \
      chmod 700 /home/vscode/.gnupg && \
      \
      echo "${GPG_PRIVATE_KEY_ASC}" > /tmp/my-gpg-key.asc && \
      \
      su vscode -c "gpg --batch --import /tmp/my-gpg-key.asc" && \
      \
      rm /tmp/my-gpg-key.asc && \
      chown -R vscode:vscode /home/vscode/.gnupg ; \
    fi

# Ensure vscode can run Docker without sudo
RUN groupadd -f docker && usermod -aG docker vscode

# Copy custom .bashrc
COPY .bashrc /home/vscode/.bashrc
RUN chown vscode:vscode /home/vscode/.bashrc

USER vscode
WORKDIR /workspaces/${LOCAL_WORKSPACE_FOLDER_BASENAME}

# Install Go‑based CLIs
RUN go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest \
    && go install github.com/tursodatabase/turso-cli/cmd/turso@latest \
    && go install github.com/seanhuebl/sqlc-qol/v2@v2.0.0 \
    && go install github.com/vektra/mockery/v2@latest \
    && go install golang.org/x/tools/cmd/goimports@latest \
    && go install github.com/pressly/goose/v3/cmd/goose@latest
