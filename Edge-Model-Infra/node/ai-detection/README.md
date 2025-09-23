# AI Detection Node - Edge-Model-Infra Implementation

This is a C++ implementation of the AI detection service using the Edge-Model-Infra framework, replacing the original Python Flask-based service.

## Overview

The AI Detection Node provides:
- Face swap detection in images and videos
- Voice synthesis detection in audio files
- Multi-modal content analysis
- High-performance inference using C++ and optimized libraries
- Integration with Edge-Model-Infra's distributed architecture

## Architecture

### Components

1. **AIDetectionNode** - Main service class inheriting from StackFlow
2. **FaceSwapDetector** - Face swap detection using OpenCV and TensorFlow C++
3. **VoiceSynthesisDetector** - Voice synthesis detection with audio processing
4. **ContentAnalyzer** - Multi-modal content analysis
5. **ModelLoader** - Generic model loading for TensorFlow/ONNX
6. **DetectionUtils** - Utility functions for file handling and processing

### Communication

- Uses ZeroMQ (pzmq) for inter-service communication
- Registers RPC actions for different detection types
- Integrates with unit-manager for API gateway functionality

## Building

### Prerequisites

- Ubuntu 20.04 or compatible Linux distribution
- CMake 3.12+
- C++17 compatible compiler
- Required libraries:
  - libzmq3-dev
  - libopencv-dev
  - libsndfile1-dev
  - libfftw3-dev
  - nlohmann-json3-dev
  - libgoogle-glog-dev
  - libboost-all-dev

### Build Instructions

```bash
cd Edge-Model-Infra/node/ai-detection
./build.sh
```

Or manually:

```bash
mkdir build && cd build
cmake .. -DCMAKE_BUILD_TYPE=Release
make -j$(nproc)
```

## Configuration

Edit `config/detection_config.json` to configure:
- Model paths
- Processing parameters
- Audio analysis settings
- Logging configuration

## Usage

### Running the Service

```bash
# Start the AI detection node
./build/ai-detection
```

### API Integration

The service integrates with the unit-manager and provides these RPC actions:

- `setup_face_detector` - Initialize face swap detection model
- `setup_voice_detector` - Initialize voice synthesis detection model
- `detect_image` - Detect face swap in image
- `detect_audio` - Detect voice synthesis in audio
- `detect_video` - Detect face swap in video
- `analyze_content` - Multi-modal content analysis
- `get_detection_status` - Get task status

### Testing

```bash
# Test the detection service
python3 test_detection.py --host localhost --port 10001
```

## Docker Deployment

### Build and Run with Docker

```bash
# Build the AI detection service
docker build -t ai-detection .

# Run with docker-compose
docker-compose -f docker-compose.ai-detection.yml up
```

### Environment Variables

- `MODEL_PATH` - Path to model files (default: /app/models)
- `UPLOAD_PATH` - Path for temporary uploads (default: /tmp/detection_uploads)

## API Compatibility

The new implementation maintains compatibility with the original Flask API:

### Detection Request
```json
{
    "request_id": "detect_001",
    "work_id": "detection",
    "action": "detect",
    "data": {
        "file_path": "/path/to/file",
        "file_type": "image"
    }
}
```

### Status Check
```json
{
    "request_id": "status_001", 
    "work_id": "detection",
    "action": "status",
    "data": {
        "task_id": "task-uuid"
    }
}
```

## Performance Improvements

Compared to the Python implementation:
- **Lower latency** - C++ native performance
- **Better memory usage** - Efficient memory management
- **Concurrent processing** - Multi-threaded inference
- **Optimized libraries** - Native OpenCV and audio processing

## Model Support

### TensorFlow Models
- Supports TensorFlow C++ API (if available)
- Frozen graph (.pb) format
- SavedModel format

### Fallback Support
- Dummy models for testing without TensorFlow
- ONNX Runtime support (planned)

## Migration from Python Service

1. **Model Conversion**: Convert Keras models to TensorFlow C++ compatible format
2. **Configuration**: Update model paths in detection_config.json
3. **Testing**: Use test_detection.py to verify functionality
4. **Deployment**: Update docker-compose to use new service

## Troubleshooting

### Common Issues

1. **Model Loading Failed**
   - Check model file paths in configuration
   - Verify TensorFlow C++ installation
   - Use dummy models for testing

2. **ZMQ Connection Issues**
   - Ensure unit-manager is running
   - Check ZMQ socket permissions
   - Verify network configuration

3. **Build Errors**
   - Install all required dependencies
   - Check CMake version compatibility
   - Verify C++17 compiler support

### Debugging

Enable debug logging by setting log level in configuration:
```json
{
    "logging": {
        "level": "DEBUG"
    }
}
```

## Future Enhancements

- ONNX Runtime integration for broader model support
- GPU acceleration with CUDA
- Real-time streaming detection
- Advanced multi-modal analysis features
- Performance monitoring and metrics
