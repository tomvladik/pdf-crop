#!/usr/bin/env bash
set -euo pipefail

ALL=false
NOCGO=false
OUT_DIR="dist"
TAGS=""

while [[ $# -gt 0 ]]; do
  case "$1" in
    --all) ALL=true; shift ;;
    --nocgo) NOCGO=true; shift ;;
    --outdir) OUT_DIR="$2"; shift 2 ;;
    --tags) TAGS="$2"; shift 2 ;;
    *) echo "Unknown argument: $1" >&2; exit 1 ;;
  esac
done

if $ALL && ! $NOCGO; then
  echo "Cross-compiling with CGO is not supported by this script. Use --nocgo or build on each target OS." >&2
  exit 1
fi

FINAL_TAGS="$TAGS"
if $NOCGO; then
  if [[ -z "$FINAL_TAGS" ]]; then
    FINAL_TAGS="nocgo"
  else
    FINAL_TAGS="$FINAL_TAGS,nocgo"
  fi
fi

mkdir -p "$OUT_DIR"

build_one() {
  local goos="$1"
  local goarch="$2"
  local ext=""
  if [[ "$goos" == "windows" ]]; then
    ext=".exe"
  fi

  GOOS="$goos" GOARCH="$goarch" CGO_ENABLED=$([[ "$NOCGO" == "true" ]] && echo 0 || echo 1) \
    go build ${FINAL_TAGS:+-tags "$FINAL_TAGS"} -o "$OUT_DIR/pdf_crop_${goos}_${goarch}${ext}" ./cmd/pdf_crop
  GOOS="$goos" GOARCH="$goarch" CGO_ENABLED=$([[ "$NOCGO" == "true" ]] && echo 0 || echo 1) \
    go build ${FINAL_TAGS:+-tags "$FINAL_TAGS"} -o "$OUT_DIR/crop_all_pdf_${goos}_${goarch}${ext}" ./cmd/crop_all_pdf
}

if $ALL; then
  build_one windows amd64
  build_one windows arm64
  build_one linux amd64
  build_one linux arm64
  build_one darwin amd64
  build_one darwin arm64
else
  build_one "$(go env GOOS)" "$(go env GOARCH)"
fi