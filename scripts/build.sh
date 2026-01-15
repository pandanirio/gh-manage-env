#!/usr/bin/env bash
set -euo pipefail

BIN="gh-manage-env"
VERSION="${1:-dev}"
OUTPUT_DIR="${OUTPUT_DIR:-dist}"

# Build for multiple platforms
PLATFORMS=(
  "darwin/amd64"
  "darwin/arm64"
  "linux/amd64"
  "linux/arm64"
)

# Create output directory
mkdir -p "$OUTPUT_DIR"

echo "Building $BIN version $VERSION"
echo "Output directory: $OUTPUT_DIR"
echo ""

# Build for each platform
for platform in "${PLATFORMS[@]}"; do
  IFS='/' read -r os arch <<< "$platform"
  
  output_name="${BIN}_${os}_${arch}"
  output_path="${OUTPUT_DIR}/${output_name}"
  
  echo "Building for $os/$arch..."
  
  GOOS="$os" GOARCH="$arch" go build \
    -ldflags "-X main.version=$VERSION" \
    -o "$output_path" \
    .
  
  # Make executable
  chmod +x "$output_path"
  
  # Show file size
  size=$(du -h "$output_path" | cut -f1)
  echo "  ✓ Built: $output_path ($size)"
done

echo ""
echo "✅ Build complete! Binaries are in $OUTPUT_DIR/"
ls -lh "$OUTPUT_DIR" | grep "$BIN" || true

