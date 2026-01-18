# Development Setup

This project uses **VS Code Dev Containers** for a consistent development environment.

## Quick Start

### Prerequisites
- **VS Code** with [Dev Containers extension](https://marketplace.visualstudio.com/items?itemName=ms-vscode-remote.remote-containers)
- **Docker Desktop** (Windows/macOS) or Docker Engine (Linux)

### Getting Started

1. Clone and open the repository:
   ```bash
   git clone <repo> pdf-crop
   cd pdf-crop
   code .
   ```

2. **Reopen in Container**:
   - Press `Ctrl+Shift+P` (or `Cmd+Shift+P` on macOS)
   - Select "Dev Containers: Reopen in Container"
   - VS Code will build and start the devcontainer

3. **Build all platforms**:
   ```bash
   make build-all
   ```

4. **Other common tasks**:
   ```bash
   make test              # Run tests
   make test-coverage     # Run tests with coverage
   make fmt               # Format code
   make vet               # Run linter
   ```

## Devcontainer Details

The devcontainer is defined in [.devcontainer/Dockerfile](.devcontainer/Dockerfile) and provides:
- **Go 1.24** (bookworm base)
- **Build essentials**: gcc, make, git
- **PDF libraries**: libmupdf-dev (for go-fitz)
- **Working directory**: `/workspaces/pdf-crop` with source mounted

### Devcontainer Settings

Defined in [.devcontainer/devcontainer.json](.devcontainer/devcontainer.json):
- **VS Code extensions**: Go language support (golang.go)
- **Post-create**: Runs `go mod download` automatically
- **Settings**: Go language server, auto-formatting

## Cross-Compilation

All cross-compilation targets (`make build-linux`, `make build-darwin`, etc.) **require the devcontainer** or a Linux/Unix environment with POSIX shell and `GOOS` environment variable support.

These targets are **not supported on Windows cmd/PowerShell** directly—use the devcontainer for cross-compilation.

## Native Builds (Current Platform Only)

To build just for your current platform (macOS/Linux native):
```bash
make build
```

This respects native CGO if available. To disable CGO:
```bash
make nocgo
```

## Troubleshooting

### Container won't start
- Ensure Docker Desktop is running
- Check the "Dev Containers" output panel in VS Code

### `make build-all` fails in native Windows shell
- Use the devcontainer: "Reopen in Container"
- Or use Git Bash / WSL from the command line

### Binaries not found in `dist/`
- Check that the build completed successfully
- Run `ls -la dist/` to see what was built

## See Also

- [Makefile](../Makefile) — All available build targets
- [.devcontainer/Dockerfile](.devcontainer/Dockerfile) — Container image definition
