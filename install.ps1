$ErrorActionPreference = "Stop"

Write-Host "Starting BaoMiHua (bmh) installation..." -ForegroundColor Cyan

# Detect architecture (assuming Windows)
$ARCH = if ($env:PROCESSOR_ARCHITECTURE -eq "ARM64") { "arm64" } else { "amd64" }
$FILE = "bmh-windows-${ARCH}.exe"
$REPO = "DeaglePC/Baomihua"

Write-Host "Detected Architecture: $ARCH"
Write-Host "Fetching latest release from $REPO..."

# Fetch latest release URL
$apiUrl = "https://api.github.com/repos/$REPO/releases/latest"
try {
    $release = Invoke-RestMethod -Uri $apiUrl -UseBasicParsing
    $asset = $release.assets | Where-Object { $_.name -eq $FILE }
    if ($null -eq $asset) {
        throw "Asset $FILE not found in the latest release."
    }
    $downloadUrl = $asset.browser_download_url
}
catch {
    Write-Host "Failed to retrieve the download URL. Exception: $_" -ForegroundColor Red
    exit 1
}

$InstallDir = "$env:LOCALAPPDATA\Programs\bmh"
if (!(Test-Path $InstallDir)) {
    New-Item -ItemType Directory -Force -Path $InstallDir | Out-Null
}

$DestFile = "$InstallDir\bmh.exe"

Write-Host "Downloading $FILE to $DestFile..."
Invoke-WebRequest -Uri $downloadUrl -OutFile $DestFile -UseBasicParsing

# Update PATH if not present
$UserPath = [Environment]::GetEnvironmentVariable("PATH", "User")
if ($UserPath -notmatch [regex]::Escape($InstallDir)) {
    Write-Host "Adding $InstallDir to user PATH..."
    $NewPath = "$UserPath;$InstallDir"
    [Environment]::SetEnvironmentVariable("PATH", $NewPath, "User")
    # Also update current process PATH so we can run it immediately
    $env:PATH += ";$InstallDir"
}

Write-Host "Installation complete." -ForegroundColor Green
Write-Host "Running initialization..." -ForegroundColor Cyan

& "$DestFile" install

