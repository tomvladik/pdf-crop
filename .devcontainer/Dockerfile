FROM golang:1.24-bookworm

# Build essentials and MuPDF dev libs for go-fitz
RUN apt-get update \
 && apt-get install -y --no-install-recommends \
    build-essential \
    git \
    pkg-config \
    libmupdf-dev \
    ca-certificates \
    make \
 && rm -rf /var/lib/apt/lists/*

# Create non-root "vscode" user for devcontainer
RUN apt-get update \
 && apt-get install -y --no-install-recommends sudo \
 && rm -rf /var/lib/apt/lists/* \
 && groupadd --gid 1000 vscode \
 && useradd -s /bin/bash --uid 1000 --gid 1000 -m vscode \
 && usermod -aG sudo vscode \
 && echo "vscode ALL=(ALL) NOPASSWD:ALL" > /etc/sudoers.d/90-vscode \
 && chmod 0440 /etc/sudoers.d/90-vscode

WORKDIR /workspaces/pdf-crop

USER vscode

# Default dev shell
CMD ["bash"]

