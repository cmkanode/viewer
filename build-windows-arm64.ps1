param(
    [string]$Output = "viewer-arm64.exe"
)

Set-StrictMode -Version Latest

$env:GOOS = "windows"
$env:GOARCH = "arm64"
$env:CGO_ENABLED = "1"

$vsWhere = "${env:ProgramFiles(x86)}\Microsoft Visual Studio\Installer\vswhere.exe"
$vsDir = $null

if (Test-Path $vsWhere) {
    $vsDir = & $vsWhere -latest -products * -requires Microsoft.VisualStudio.Component.VC.Tools.x86.x64 -property installationPath
    if ($LASTEXITCODE -eq 0 -and $vsDir) {
        $vcTools = Join-Path $vsDir "VC\Tools\MSVC"
        if (Test-Path $vcTools) {
            $msvcVersion = Get-ChildItem $vcTools | Sort-Object Name -Descending | Select-Object -First 1
            if ($msvcVersion) {
                $vcBin = Join-Path $msvcVersion.FullName "bin\Hostx64\x64"
                $vcVars = Join-Path $vcBin "cl.exe"
                if (Test-Path $vcVars) {
                    $env:CC = $vcVars
                    $env:CXX = $vcVars
                    $env:PATH = "$vcBin;$env:PATH"
                }
            }
        }
    }
}

if (-not $env:CC) {
    $env:CC = "aarch64-w64-mingw32-gcc"
}
if (-not $env:CXX) {
    $env:CXX = "aarch64-w64-mingw32-g++"
}

if (-not (Get-Command $env:CC -ErrorAction SilentlyContinue)) {
    Write-Error "ARM64 Windows compiler not found. Install Visual Studio Build Tools with ARM64 support or a MinGW-w64 ARM64 toolchain, then rerun this script."
    exit 1
}

Write-Host "Building viewer for Windows ARM64 -> $Output"

go build -ldflags="-H=windowsgui" -o $Output .

if ($LASTEXITCODE -ne 0) {
    Write-Error "go build failed with exit code $LASTEXITCODE"
    exit $LASTEXITCODE
}

Write-Host "Build complete: $Output"
