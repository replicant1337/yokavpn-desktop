#!/bin/bash
# Release script - builds and uploads to GitHub

VERSION=${1:-$(date +%v%Y.%m.%d)}
REPO="replicant1337/yokavpn-desktop"

echo "Building YokaVPN v$VERSION..."

# Build Windows
cd /d E:\Devlop\xray-client
wails build -ldflags "-H=windowsgui"

# Copy to releases
copy "build\bin\YokaVPN.exe" "build\releases\"

# Commit and tag
cd build\releases
git add YokaVPN.exe
git commit -m "YokaVPN v$VERSION"
git tag -a "v$VERSION" -m "YokaVPN v$VERSION"

# Push
git push
git push origin "v$VERSION"

echo "Done! Release v$VERSION created."
