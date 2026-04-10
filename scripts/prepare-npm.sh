#!/bin/bash
set -e

VERSION="${GITHUB_REF_NAME:-v0.1.0}"
VERSION="${VERSION#v}"

echo "Preparing npm packages v$VERSION"

cd npm/gitdig
npm version "$VERSION" --no-git-tag-version --allow-same-version
cd ../..

declare -A PLATFORMS=(
  ["darwin-arm64"]="darwin_arm64"
  ["darwin-x64"]="darwin_amd64"
  ["linux-arm64"]="linux_arm64"
  ["linux-x64"]="linux_amd64"
  ["win32-arm64"]="windows_arm64"
  ["win32-x64"]="windows_amd64"
)

for npm_platform in "${!PLATFORMS[@]}"; do
  goreleaser_platform="${PLATFORMS[$npm_platform]}"
  pkg_dir="npm/@gitdig/$npm_platform"

  echo "Creating $npm_platform..."

  mkdir -p "$pkg_dir/bin"

  if [[ "$npm_platform" == win32* ]]; then
    binary_name="gitdig.exe"
  else
    binary_name="gitdig"
  fi

  binary_src=$(find dist -name "$binary_name" -path "*${goreleaser_platform}*" -type f | head -1)

  if [ -n "$binary_src" ]; then
    cp "$binary_src" "$pkg_dir/bin/$binary_name"
    chmod +x "$pkg_dir/bin/$binary_name"
  fi

  os="${npm_platform%-*}"
  cpu="${npm_platform#*-}"

  cat > "$pkg_dir/package.json" <<EOF
{
  "name": "@gitdig/$npm_platform",
  "version": "$VERSION",
  "description": "GitDig binary for $npm_platform",
  "repository": "https://github.com/rishabnotfound/GitDig",
  "license": "MIT",
  "os": ["$os"],
  "cpu": ["$cpu"],
  "files": ["bin"]
}
EOF

done

echo "Done"
