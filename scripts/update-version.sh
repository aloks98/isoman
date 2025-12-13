#!/bin/bash
set -e

VERSION=$1

if [ -z "$VERSION" ]; then
  echo "Usage: $0 <version>"
  echo "Example: $0 1.2.3"
  exit 1
fi

# Validate version format (semver: X.Y.Z)
if ! [[ "$VERSION" =~ ^[0-9]+\.[0-9]+\.[0-9]+$ ]]; then
  echo "Error: Version must be in semver format (X.Y.Z)"
  echo "Example: 1.2.3"
  exit 1
fi

echo "Updating version to $VERSION..."

# Update root package.json
if [ -f "package.json" ]; then
  echo "Updating root package.json..."
  sed -i.bak "s/\"version\": \".*\"/\"version\": \"$VERSION\"/" package.json
  rm -f package.json.bak
fi

# Update UI package.json
if [ -f "ui/package.json" ]; then
  echo "Updating ui/package.json..."
  sed -i.bak "s/\"version\": \".*\"/\"version\": \"$VERSION\"/" ui/package.json
  rm -f ui/package.json.bak
fi

# Create VERSION file for backend
echo "$VERSION" > VERSION

echo "✅ Version updated to $VERSION"
echo ""
echo "Files updated:"
echo "  - package.json"
echo "  - ui/package.json"
echo "  - VERSION"
