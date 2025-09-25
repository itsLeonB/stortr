#!/bin/sh

# Pre-push hook that runs linting and tests before allowing push

echo "Running pre-push checks..."

# Run linting
echo "\n=== Running linting ==="
if ! make lint; then
    echo "\n❌ Linting failed! Please fix the lint issues before pushing.\n"
    exit 1
fi

# Run tests
echo "\n=== Running tests ==="
if ! make test; then
    echo "\n❌ Tests failed! Please fix the test issues before pushing.\n"
    exit 1
fi

# Run build
echo "\n=== Running build ==="
if ! make build-grpc; then
    echo "\n❌ Build failed! Please fix the build issues before pushing.\n"
    exit 1
fi
if ! make build-job; then
    echo "\n❌ Build failed! Please fix the build issues before pushing.\n"
    exit 1
fi

echo "\n✅ All checks passed! Pushing can continue...\n"
