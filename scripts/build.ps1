param(
    [switch]$All,
    [switch]$NoCgo,
    [string]$OutDir = "dist",
    [string]$Tags = ""
)

$ErrorActionPreference = "Stop"

if ($All -and -not $NoCgo) {
    Write-Error "Cross-compiling with CGO is not supported by this script. Use -NoCgo (purego) or build on each target OS."
    exit 1
}

$finalTags = $Tags
if ($NoCgo) {
    if ([string]::IsNullOrWhiteSpace($finalTags)) {
        $finalTags = "nocgo"
    } else {
        $finalTags = "$finalTags,nocgo"
    }
}

if (!(Test-Path $OutDir)) {
    New-Item -ItemType Directory -Path $OutDir | Out-Null
}

function Build-One {
    param(
        [string]$GOOS,
        [string]$GOARCH
    )

    $env:GOOS = $GOOS
    $env:GOARCH = $GOARCH
    $env:CGO_ENABLED = $(if ($NoCgo) { "0" } else { "1" })

    $ext = $(if ($GOOS -eq "windows") { ".exe" } else { "" })
    $pdfCropOut = Join-Path $OutDir "pdf_crop_${GOOS}_${GOARCH}${ext}"
    $cropAllOut = Join-Path $OutDir "crop_all_pdf_${GOOS}_${GOARCH}${ext}"

    if ([string]::IsNullOrWhiteSpace($finalTags)) {
        & go build -o $pdfCropOut .\cmd\pdf_crop
        & go build -o $cropAllOut .\cmd\crop_all_pdf
    } else {
        & go build -tags $finalTags -o $pdfCropOut .\cmd\pdf_crop
        & go build -tags $finalTags -o $cropAllOut .\cmd\crop_all_pdf
    }
}

if ($All) {
    $targets = @(
        @{ os = "windows"; arch = "amd64" },
        @{ os = "windows"; arch = "arm64" },
        @{ os = "linux"; arch = "amd64" },
        @{ os = "linux"; arch = "arm64" },
        @{ os = "darwin"; arch = "amd64" },
        @{ os = "darwin"; arch = "arm64" }
    )

    foreach ($t in $targets) {
        Build-One -GOOS $t.os -GOARCH $t.arch
    }
} else {
    $goos = $(go env GOOS)
    $goarch = $(go env GOARCH)
    Build-One -GOOS $goos -GOARCH $goarch
}