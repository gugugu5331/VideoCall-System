#!/bin/bash

# Build script for AI Detection Node
# This script builds the AI detection service using Edge-Model-Infra

set -e

echo "Building AI Detection Node..."

# Check if we're in the right directory
if [ ! -f "CMakeLists.txt" ]; then
    echo "Error: CMakeLists.txt not found. Please run this script from the ai-detection directory."
    exit 1
fi

# Create build directory
mkdir -p build
cd build

# Configure with CMake
echo "Configuring with CMake..."
cmake .. \
    -DCMAKE_BUILD_TYPE=Release \
    -DCMAKE_CXX_STANDARD=17

# Build the project
echo "Building..."
make -j$(nproc)

echo "Build completed successfully!"
echo "Executable: $(pwd)/ai-detection"

# Check if the executable was created
if [ -f "ai-detection" ]; then
    echo "AI Detection Node built successfully!"
    echo "To run: ./build/ai-detection"
else
    echo "Error: Build failed - executable not found"
    exit 1
fi
