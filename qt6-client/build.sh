#!/bin/bash

# Qt6 Meeting System Client Build Script

set -e

echo "================================"
echo "Qt6 Meeting System Client Build"
echo "================================"

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Check Qt6 installation
echo -e "${YELLOW}Checking Qt6 installation...${NC}"
if ! command -v qmake6 &> /dev/null && ! command -v qmake &> /dev/null; then
    echo -e "${RED}Error: Qt6 is not installed or not in PATH${NC}"
    echo "Please install Qt6 from https://www.qt.io/download"
    exit 1
fi

echo -e "${GREEN}Qt6 found${NC}"

# Check CMake
echo -e "${YELLOW}Checking CMake installation...${NC}"
if ! command -v cmake &> /dev/null; then
    echo -e "${RED}Error: CMake is not installed${NC}"
    echo "Please install CMake 3.16 or higher"
    exit 1
fi

CMAKE_VERSION=$(cmake --version | head -n1 | cut -d' ' -f3)
echo -e "${GREEN}CMake $CMAKE_VERSION found${NC}"

# Create build directory
BUILD_DIR="build"
if [ -d "$BUILD_DIR" ]; then
    echo -e "${YELLOW}Removing existing build directory...${NC}"
    rm -rf "$BUILD_DIR"
fi

echo -e "${YELLOW}Creating build directory...${NC}"
mkdir -p "$BUILD_DIR"
cd "$BUILD_DIR"

# Configure
echo -e "${YELLOW}Configuring project...${NC}"
cmake .. -DCMAKE_BUILD_TYPE=Release

# Build
echo -e "${YELLOW}Building project...${NC}"
cmake --build . --config Release -j$(nproc 2>/dev/null || sysctl -n hw.ncpu 2>/dev/null || echo 4)

# Check if build succeeded
if [ $? -eq 0 ]; then
    echo -e "${GREEN}================================${NC}"
    echo -e "${GREEN}Build completed successfully!${NC}"
    echo -e "${GREEN}================================${NC}"
    echo ""
    echo "Executable location: $BUILD_DIR/bin/MeetingSystemClient"
    echo ""
    echo "To run the application:"
    echo "  cd $BUILD_DIR/bin"
    echo "  ./MeetingSystemClient"
else
    echo -e "${RED}================================${NC}"
    echo -e "${RED}Build failed!${NC}"
    echo -e "${RED}================================${NC}"
    exit 1
fi

