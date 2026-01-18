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

### Windows Runtime Setup (Purego/No-CGO builds)

When using a Windows binary built from Linux via `make build-windows`:

1. **Download MuPDF libraries** matching your binary architecture:
   - 64-bit: Download `mupdf-X.Y-windows-x64.zip` from [MuPDF releases](https://mupdf.com/releases)
   - 32-bit: Download `mupdf-X.Y-windows-x32.zip`

2. **Install to system PATH**:
   - Extract and add the directory containing `mupdf.dll` to your Windows `PATH`
   - Or place `mupdf.dll` in the same directory as the binary

3. **Install libffi** (for FFI bindings):
   - Download libffi from [GitHub releases](https://github.com/winehq/wine/tree/master/libs/wine)
   - Or install via package manager (e.g., `choco install libffi` on Windows with Chocolatey)
   - Ensure `libffi.dll` is in `PATH` or same directory as binary

**Note**: Verify the MuPDF version matches what the binary was built against. You can check the MuPDF version in the go-fitz dependency in `go.mod`.

## Usage

### pdf_crop

```
pdf_crop -i input.pdf --threshold 0.008 --space 5
pdf_crop -i input.pdf -p 0 0 0 0 0 out0.pdf
pdf_crop --help
```

### crop_all_pdf

```
crop_all_pdf --dir ./pdfs --threshold 0.1
crop_all_pdf --help
```

## Library usage

Import the package and call the crop helpers directly. Example: crop every page and write the cropped pages back into a single (multi-page) PDF, using defaults plus a bit of extra whitespace.

```go
package main

import (
  "log"

  "pdf-crop/pkg/crop"
)

func main() {
  opts := crop.DefaultOptions()
  opts.Space = 8 // add extra points of whitespace

  results, err := crop.CropAllPagesToSingleFile("input.pdf", "output.pdf", opts)
  if err != nil {
    log.Fatalf("crop: %v", err)
  }

  for _, r := range results {
    log.Printf("page %d => %s (media %s)", r.PageNo, r.Output, crop.RectString(r.Media))
  }
}
```

## Page Size Fallback

- When a page's `MediaBox` is missing or page boundaries cannot be read, cropping falls back to A4 dimensions: 595 × 842 points.
- This fallback only influences the computed crop rectangle; existing PDFs with valid page sizes are used as-is.
- If you need a different default, ensure your input PDF defines `MediaBox` for all pages or pre-process it to set page boundaries.

## License

Project license: AGPL-3.0. See [LICENSE](LICENSE).
Uses `pdfcpu` (Apache-2.0).