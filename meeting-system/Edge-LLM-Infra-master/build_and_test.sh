#!/bin/bash

set -e

echo "=========================================="
echo "AI Inference System Build and Test Script"
echo "=========================================="

SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
cd "$SCRIPT_DIR"

# Step 1: Build Docker image
echo ""
echo "[Step 1] Building Docker image..."
echo "=========================================="
cd docker/build
if docker build -t llm:v1.0 -f base.dockerfile .; then
    echo "✓ Docker image built successfully"
else
    echo "✗ Docker image build failed"
    exit 1
fi

# Step 2: Verify Docker image
echo ""
echo "[Step 2] Verifying Docker image..."
echo "=========================================="
if docker images | grep -q "llm.*v1.0"; then
    echo "✓ Docker image verified"
else
    echo "✗ Docker image not found"
    exit 1
fi

# Step 3: Build infra-controller
echo ""
echo "[Step 3] Building infra-controller..."
echo "=========================================="
cd "$SCRIPT_DIR"
mkdir -p install/lib
cd infra-controller
mkdir -p build && cd build
if cmake .. && make -j12; then
    echo "✓ infra-controller built successfully"
else
    echo "✗ infra-controller build failed"
    exit 1
fi

# Step 4: Build unit-manager
echo ""
echo "[Step 4] Building unit-manager..."
echo "=========================================="
cd "$SCRIPT_DIR/unit-manager"
mkdir -p build && cd build
if cmake .. && make -j12; then
    echo "✓ unit-manager built successfully"
else
    echo "✗ unit-manager build failed"
    exit 1
fi

echo ""
echo "=========================================="
echo "✓ Build completed successfully!"
echo "=========================================="
echo ""
echo "Next steps:"
echo "1. Start unit_manager:"
echo "   cd $SCRIPT_DIR/unit-manager/build"
echo "   ./unit_manager"
echo ""
echo "2. In another terminal, start Docker container:"
echo "   cd $SCRIPT_DIR/docker/scripts"
echo "   bash llm_docker_run.sh"
echo ""
echo "3. Enter Docker container and build AI inference node:"
echo "   bash llm_docker_into.sh"
echo "   cd /work/node/llm"
echo "   mkdir build && cd build"
echo "   cmake .. && make -j12"
echo "   ./llm"
echo ""
echo "4. In another terminal, run tests:"
echo "   cd $SCRIPT_DIR/node/llm"
echo "   python3 test_ai_inference.py"
echo ""

