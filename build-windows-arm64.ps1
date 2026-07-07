param(
    [string]$Output = "viewer-arm64.exe"
)

Set-StrictMode -Version Latest

$env:GOOS = "windows"
$env:GOARCH = "arm64"

Write-Host "Building viewer for Windows ARM64 -> $Output"

go build -ldflags="-H=windowsgui" -o $Output .

if ($LASTEXITCODE -ne 0) {
    Write-Error "go build failed with exit code $LASTEXITCODE"
    exit $LASTEXITCODE
}

Write-Host "Build complete: $Output"
