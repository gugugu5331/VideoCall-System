#!/bin/bash

set -e

echo "Installing ONNX Runtime..."

# Download ONNX Runtime
ONNX_VERSION="1.16.3"
ONNX_FILE="onnxruntime-linux-x64-${ONNX_VERSION}.tgz"
ONNX_URL="https://github.com/microsoft/onnxruntime/releases/download/v${ONNX_VERSION}/${ONNX_FILE}"

cd /tmp

if [ ! -f "${ONNX_FILE}" ]; then
    echo "Downloading ONNX Runtime ${ONNX_VERSION}..."
    wget ${ONNX_URL}
fi

echo "Extracting ONNX Runtime..."
tar -xzf ${ONNX_FILE}

echo "Installing ONNX Runtime headers and libraries..."
ONNX_DIR="onnxruntime-linux-x64-${ONNX_VERSION}"
cp -r ${ONNX_DIR}/include/* /usr/local/include/
cp -r ${ONNX_DIR}/lib/* /usr/local/lib/

echo "Updating library cache..."
ldconfig

echo "Cleaning up..."
rm -rf ${ONNX_FILE} ${ONNX_DIR}

echo "ONNX Runtime installed successfully!"

