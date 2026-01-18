# pdf-crop

Go reimplementation of the Python `pdf_crop` workflow with an exact raster-based detection approach.

## Binaries

Two CLIs are provided:

- `pdf_crop`: crops pages using raster detection and writes per-page PDFs (matches the Python tool’s behavior).
- `crop_all_pdf`: processes all PDFs in a directory and writes one cropped PDF per input.

## Build

Use the Makefile for all builds.

- Current platform (CGO by default):
  - `make build`
- No CGO (purego mode):
  - `make nocgo`
- Cross-compile (Linux/macOS/Windows, nocgo):
  - `make build-all`
  - Or specific targets: `make build-linux`, `make build-linux-arm64`, `make build-darwin`, `make build-darwin-arm64`, `make build-windows`, `make build-windows-arm64`

Outputs go to `dist/` by default (override with `DIST_DIR=...`).

Windows note: Use GNU Make via Git Bash, MSYS2, or WSL. For CGO builds ensure MSVC Build Tools and MuPDF dev libraries are installed; otherwise use `make nocgo`.

## Runtime dependencies (CGO builds)

The raster renderer uses MuPDF via go-fitz.

### Windows

Recommended: build on Windows with MSVC Build Tools and use the bundled library in go-fitz.

### Linux

Install MuPDF development libraries (for example, `libmupdf-dev`) and a C toolchain.

### macOS

Install MuPDF via Homebrew and ensure clang is available.

## Runtime dependencies (purego `--nocgo`)

Purego mode still requires MuPDF shared libraries and libffi at runtime. Set the exact MuPDF version with `FZ_VERSION` (or set `fitz.FzVersion` in code) to match the installed library.

## Usage

### pdf_crop

```
pdf_crop -i input.pdf --threshold 0.008 --space 5
pdf_crop -i input.pdf -p 0 0 0 0 0 out0.pdf
```

### crop_all_pdf

```
crop_all_pdf --dir ./pdfs --threshold 0.1
```

## License

Project license: AGPL-3.0. See [LICENSE](LICENSE).
Uses `pdfcpu` (Apache-2.0).