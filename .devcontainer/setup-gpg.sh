#!/usr/bin/env bash
set -euxo pipefail

# Unset any local overrides
unset GIT_AUTHOR_EMAIL || true
unset GIT_COMMITTER_EMAIL || true
git config --local --unset-all user.email || true

# Import a mounted GPG key if it exists
if [ -f /home/vscode/my-gpg-key.asc ]; then
  mkdir -p /home/vscode/.gnupg
  chmod 700 /home/vscode/.gnupg
  gpg --batch --import /home/vscode/my-gpg-key.asc
  rm /home/vscode/my-gpg-key.asc
  chown -R vscode:vscode /home/vscode/.gnupg
fi

# Fix ownership if $GPG_PRIVATE_KEY_ASC was already imported by Dockerfile
if [ -n "${GPG_PRIVATE_KEY_ASC:-}" ]; then
  chown -R vscode:vscode /home/vscode/.gnupg
fi

# Configure global Git identity and GPG
git config --global user.name "sean huebl"
git config --global user.email "sean.huebl@gmail.com"
git config --global user.signingkey "3576E0EB31EDA666"
git config --global commit.gpgsign true
git config --global --unset gpg.program || true
git config --global gpg.program gpg