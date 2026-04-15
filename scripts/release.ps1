# YokaVPN Automatic Release Script
# Requirements: GitHub CLI
# Install: winget install GitHub.cli
# Then: gh auth login
# Run: powershell -File scripts\release.ps1 -Version "1.0.7"

param(
    [string]$Version = ""
)

# Check if gh is installed
$ghPath = Get-Command gh -ErrorAction SilentlyContinue
if (-not $ghPath) {
    Write-Host "Installing GitHub CLI..." -ForegroundColor Yellow
    winget install GitHub.cli --accept-package-agreements --accept-source-agreements
    
    Write-Host "Please run: gh auth login" -ForegroundColor Yellow
    Write-Host "Then run this script again." -ForegroundColor Green
    exit 1
}

if ($Version -eq "") {
    $Version = Get-Date -Format "yyyy.MM.dd"
}

$REPO = "replicant1337/yokavpn-desktop"
$APP_DIR = "E:\Devlop\xray-client"

Write-Host "Building YokaVPN v$Version..." -ForegroundColor Cyan

# Build
Set-Location $APP_DIR
wails build -ldflags "-H=windowsgui"

if (-not $?) { 
    Write-Host "Build failed!" -ForegroundColor Red
    exit 1
}

# Copy to releases
Copy-Item "build\bin\YokaVPN.exe" "build\releases\"

# Commit and tag
Set-Location "build\releases"
git add YokaVPN.exe
git commit -m "YokaVPN v$Version"
git tag -a "v$Version" -m "YokaVPN v$Version"

# Push commits
Write-Host "Pushing to GitHub..." -ForegroundColor Yellow
git push origin master
git push origin "v$Version"

# Wait for push
Start-Sleep -Seconds 3

# Create release using GitHub CLI
Write-Host "Creating GitHub release..." -ForegroundColor Yellow
gh release create "v$Version" `
    --title "YokaVPN v$Version" `
    --generate-notes `
    --target master

# Upload Windows exe
Write-Host "Uploading Windows exe..." -ForegroundColor Yellow
gh release upload "v$Version" "build\releases\YokaVPN.exe" --clobber

Write-Host ""
Write-Host "===============================================" -ForegroundColor Green
Write-Host "Windows release created!" -ForegroundColor Green
Write-Host "Now building for other platforms..." -ForegroundColor Yellow
Write-Host "===============================================" -ForegroundColor Green

# For multi-platform builds, push tag to trigger GitHub Actions
Write-Host ""
Write-Host "GitHub Actions will build Linux and macOS automatically." -ForegroundColor Cyan
Write-Host "Check: https://github.com/$REPO/actions" -ForegroundColor Cyan
