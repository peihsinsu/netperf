# PowerShell build script for Windows users
# Cross-platform build for bandfetch

param(
    [string]$Version = "dev"
)

$ErrorActionPreference = "Stop"

if ($Version -eq "dev") {
    try {
        $Version = git describe --tags --always --dirty 2>$null
    } catch {
        $Version = "dev"
    }
}

$Binary = "bandfetch"
$BuildDir = "bin"
$ReleaseDir = "releases\$Version"

Write-Host "╔════════════════════════════════════════════════════════╗" -ForegroundColor Cyan
Write-Host "║     bandfetch Cross-Platform Build Script             ║" -ForegroundColor Cyan
Write-Host "║     Version: $Version" -ForegroundColor Cyan
Write-Host "╚════════════════════════════════════════════════════════╝" -ForegroundColor Cyan
Write-Host ""

# Clean and create directories
Write-Host "→ Preparing directories..." -ForegroundColor Yellow
if (Test-Path $BuildDir) { Remove-Item -Recurse -Force $BuildDir }
if (Test-Path $ReleaseDir) { Remove-Item -Recurse -Force $ReleaseDir }
New-Item -ItemType Directory -Force -Path $BuildDir | Out-Null
New-Item -ItemType Directory -Force -Path $ReleaseDir | Out-Null
Write-Host "  ✓ Directories ready" -ForegroundColor Green
Write-Host ""

Write-Host "→ Building for all platforms..." -ForegroundColor Yellow
Write-Host ""

$builds = @(
    @{Name="Linux (AMD64)"; GOOS="linux"; GOARCH="amd64"; Output="$Binary-linux-amd64"},
    @{Name="Linux (ARM64)"; GOOS="linux"; GOARCH="arm64"; Output="$Binary-linux-arm64"},
    @{Name="Windows (AMD64)"; GOOS="windows"; GOARCH="amd64"; Output="$Binary-windows-amd64.exe"},
    @{Name="Windows (ARM64)"; GOOS="windows"; GOARCH="arm64"; Output="$Binary-windows-arm64.exe"},
    @{Name="macOS (AMD64)"; GOOS="darwin"; GOARCH="amd64"; Output="$Binary-darwin-amd64"},
    @{Name="macOS (ARM64)"; GOOS="darwin"; GOARCH="arm64"; Output="$Binary-darwin-arm64"}
)

$i = 1
foreach ($build in $builds) {
    Write-Host "  [$i/$($builds.Count)] $($build.Name)..." -ForegroundColor Cyan

    $env:GOOS = $build.GOOS
    $env:GOARCH = $build.GOARCH
    $env:CGO_ENABLED = "0"

    $ldflags = "-s -w -X main.version=$Version"
    $outputPath = Join-Path $BuildDir $build.Output

    & go build -ldflags $ldflags -o $outputPath .\cmd\bandfetch

    if ($LASTEXITCODE -eq 0) {
        Write-Host "        ✓ Complete" -ForegroundColor Green
    } else {
        Write-Host "        ✗ Failed" -ForegroundColor Red
        exit 1
    }
    $i++
}

Write-Host ""
Write-Host "→ Creating release archives..." -ForegroundColor Yellow
Write-Host ""

function Create-Archive {
    param($Platform, $BinaryFile)

    $archiveName = "$Binary-$Version-$Platform"
    $binaryPath = Join-Path $BuildDir $BinaryFile

    if ($BinaryFile -match "\.exe$") {
        # Windows: zip archive
        $zipPath = Join-Path $ReleaseDir "$archiveName.zip"
        Compress-Archive -Path $binaryPath -DestinationPath $zipPath -Force
        Write-Host "  ✓ $archiveName.zip" -ForegroundColor Green
    } else {
        # Unix: tar.gz (requires tar.exe in Windows 10+)
        Push-Location $BuildDir
        $tarPath = Join-Path ".." $ReleaseDir "$archiveName.tar.gz"
        & tar -czf $tarPath $BinaryFile
        Pop-Location
        Write-Host "  ✓ $archiveName.tar.gz" -ForegroundColor Green
    }
}

Create-Archive "linux-amd64" "$Binary-linux-amd64"
Create-Archive "linux-arm64" "$Binary-linux-arm64"
Create-Archive "windows-amd64" "$Binary-windows-amd64.exe"
Create-Archive "windows-arm64" "$Binary-windows-arm64.exe"
Create-Archive "darwin-amd64" "$Binary-darwin-amd64"
Create-Archive "darwin-arm64" "$Binary-darwin-arm64"

Write-Host ""
Write-Host "→ Generating checksums..." -ForegroundColor Yellow
Push-Location $ReleaseDir
$files = Get-ChildItem -File
$checksums = $files | ForEach-Object {
    $hash = (Get-FileHash $_.Name -Algorithm SHA256).Hash.ToLower()
    "$hash  $($_.Name)"
}
$checksums | Out-File -FilePath "SHA256SUMS" -Encoding ASCII
Write-Host "  ✓ SHA256SUMS created" -ForegroundColor Green
Pop-Location

Write-Host ""
Write-Host "╔════════════════════════════════════════════════════════╗" -ForegroundColor Cyan
Write-Host "║              Build Complete!                           ║" -ForegroundColor Cyan
Write-Host "╚════════════════════════════════════════════════════════╝" -ForegroundColor Cyan
Write-Host ""
Write-Host "Release artifacts:" -ForegroundColor Yellow
Get-ChildItem $ReleaseDir | Format-Table -AutoSize
Write-Host ""
Write-Host "Archives ready for distribution in: $ReleaseDir\" -ForegroundColor Green
