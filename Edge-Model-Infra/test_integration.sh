#!/bin/bash

# Integration test script for AI Detection Node
# Tests the complete Edge-Model-Infra based AI detection system

set -e

echo "=== AI Detection Node Integration Test ==="

# Configuration
UNIT_MANAGER_HOST="localhost"
UNIT_MANAGER_PORT="10001"
TEST_DIR="$(dirname "$0")"
AI_DETECTION_DIR="$TEST_DIR/node/ai-detection"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Function to print colored output
print_status() {
    local status=$1
    local message=$2
    case $status in
        "INFO")
            echo -e "${YELLOW}[INFO]${NC} $message"
            ;;
        "SUCCESS")
            echo -e "${GREEN}[SUCCESS]${NC} $message"
            ;;
        "ERROR")
            echo -e "${RED}[ERROR]${NC} $message"
            ;;
    esac
}

# Function to check if service is running
check_service() {
    local service_name=$1
    local port=$2
    
    if nc -z localhost $port 2>/dev/null; then
        print_status "SUCCESS" "$service_name is running on port $port"
        return 0
    else
        print_status "ERROR" "$service_name is not running on port $port"
        return 1
    fi
}

# Function to build AI detection node
build_ai_detection() {
    print_status "INFO" "Building AI Detection Node..."
    
    cd "$AI_DETECTION_DIR"
    
    if [ ! -f "build.sh" ]; then
        print_status "ERROR" "build.sh not found in $AI_DETECTION_DIR"
        return 1
    fi
    
    if ./build.sh; then
        print_status "SUCCESS" "AI Detection Node built successfully"
        return 0
    else
        print_status "ERROR" "Failed to build AI Detection Node"
        return 1
    fi
}

# Function to start services
start_services() {
    print_status "INFO" "Starting services..."
    
    # Start unit-manager in background
    cd "$TEST_DIR/unit-manager"
    if [ -f "build/unit-manager" ]; then
        print_status "INFO" "Starting unit-manager..."
        ./build/unit-manager &
        UNIT_MANAGER_PID=$!
        sleep 3
    else
        print_status "ERROR" "unit-manager executable not found"
        return 1
    fi
    
    # Start AI detection node in background
    cd "$AI_DETECTION_DIR"
    if [ -f "build/ai-detection" ]; then
        print_status "INFO" "Starting AI detection node..."
        ./build/ai-detection &
        AI_DETECTION_PID=$!
        sleep 3
    else
        print_status "ERROR" "ai-detection executable not found"
        return 1
    fi
    
    return 0
}

# Function to stop services
stop_services() {
    print_status "INFO" "Stopping services..."
    
    if [ ! -z "$AI_DETECTION_PID" ]; then
        kill $AI_DETECTION_PID 2>/dev/null || true
        print_status "INFO" "AI detection node stopped"
    fi
    
    if [ ! -z "$UNIT_MANAGER_PID" ]; then
        kill $UNIT_MANAGER_PID 2>/dev/null || true
        print_status "INFO" "Unit manager stopped"
    fi
}

# Function to run Python tests
run_python_tests() {
    print_status "INFO" "Running Python integration tests..."
    
    cd "$AI_DETECTION_DIR"
    
    if [ -f "test_detection.py" ]; then
        if python3 test_detection.py --host $UNIT_MANAGER_HOST --port $UNIT_MANAGER_PORT; then
            print_status "SUCCESS" "Python tests passed"
            return 0
        else
            print_status "ERROR" "Python tests failed"
            return 1
        fi
    else
        print_status "ERROR" "test_detection.py not found"
        return 1
    fi
}

# Function to test basic connectivity
test_connectivity() {
    print_status "INFO" "Testing connectivity..."
    
    # Test unit-manager
    if check_service "Unit Manager" $UNIT_MANAGER_PORT; then
        print_status "SUCCESS" "Unit Manager connectivity OK"
    else
        print_status "ERROR" "Cannot connect to Unit Manager"
        return 1
    fi
    
    return 0
}

# Function to run manual API tests
test_api_manually() {
    print_status "INFO" "Running manual API tests..."
    
    # Test with curl if available
    if command -v curl &> /dev/null; then
        print_status "INFO" "Testing with curl..."
        
        # Create test JSON
        local test_json='{"request_id":"test_001","work_id":"detection","action":"setup","data":{"detector_type":"face_swap","model_path":"/tmp/test_model.pb"}}'
        
        # Note: This is a simplified test - the actual protocol is TCP with newlines
        print_status "INFO" "API test would require TCP client, skipping curl test"
    else
        print_status "INFO" "curl not available, skipping manual API test"
    fi
    
    return 0
}

# Cleanup function
cleanup() {
    print_status "INFO" "Cleaning up..."
    stop_services
    cd "$TEST_DIR"
}

# Set trap for cleanup
trap cleanup EXIT

# Main test sequence
main() {
    print_status "INFO" "Starting integration test sequence..."
    
    # Check prerequisites
    print_status "INFO" "Checking prerequisites..."
    
    if ! command -v cmake &> /dev/null; then
        print_status "ERROR" "cmake is required but not installed"
        exit 1
    fi
    
    if ! command -v python3 &> /dev/null; then
        print_status "ERROR" "python3 is required but not installed"
        exit 1
    fi
    
    # Build AI detection node
    if ! build_ai_detection; then
        print_status "ERROR" "Build failed"
        exit 1
    fi
    
    # Start services
    if ! start_services; then
        print_status "ERROR" "Failed to start services"
        exit 1
    fi
    
    # Wait for services to initialize
    print_status "INFO" "Waiting for services to initialize..."
    sleep 5
    
    # Test connectivity
    if ! test_connectivity; then
        print_status "ERROR" "Connectivity test failed"
        exit 1
    fi
    
    # Run Python tests
    if ! run_python_tests; then
        print_status "ERROR" "Python tests failed"
        exit 1
    fi
    
    # Run manual API tests
    test_api_manually
    
    print_status "SUCCESS" "All integration tests completed successfully!"
    
    return 0
}

# Run main function
main "$@"
