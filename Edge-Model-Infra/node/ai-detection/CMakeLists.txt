cmake_minimum_required(VERSION 3.12)
project(ai-detection)

set(CMAKE_CXX_STANDARD 17)
set(CMAKE_CXX_STANDARD_REQUIRED ON)

find_package(Threads REQUIRED)
find_package(OpenCV REQUIRED)
find_package(PkgConfig REQUIRED)

# Find TensorFlow C++ (if available)
# Note: TensorFlow C++ installation may vary by system
# Alternative: Use ONNX Runtime for model inference
find_library(TENSORFLOW_LIB tensorflow_cc PATHS /usr/local/lib /usr/lib)
find_library(TENSORFLOW_FRAMEWORK_LIB tensorflow_framework PATHS /usr/local/lib /usr/lib)

# Audio processing libraries
pkg_check_modules(SNDFILE REQUIRED sndfile)
pkg_check_modules(FFTW3 REQUIRED fftw3)

include_directories(
    ${CMAKE_SOURCE_DIR}/include
    ../../hybrid-comm/include
    ../../infra-controller/include
    ../../utils
    ${OpenCV_INCLUDE_DIRS}
    ${SNDFILE_INCLUDE_DIRS}
    ${FFTW3_INCLUDE_DIRS}
    /usr/include/nlohmann
)

# Source files
file(GLOB SRCS 
    src/main.cpp
    src/ai_detection_node.cpp
    src/face_swap_detector.cpp
    src/voice_synthesis_detector.cpp
    src/content_analyzer.cpp
    src/model_loader.cpp
    src/detection_utils.cpp
)

add_executable(${PROJECT_NAME} ${SRCS})

target_link_libraries(${PROJECT_NAME}
    PRIVATE
    pthread
    ${CMAKE_SOURCE_DIR}/../../install/lib/libstackflow.a
    zmq
    ${OpenCV_LIBS}
    ${SNDFILE_LIBRARIES}
    ${FFTW3_LIBRARIES}
)

# Link TensorFlow if available
if(TENSORFLOW_LIB AND TENSORFLOW_FRAMEWORK_LIB)
    target_link_libraries(${PROJECT_NAME} PRIVATE ${TENSORFLOW_LIB} ${TENSORFLOW_FRAMEWORK_LIB})
    target_compile_definitions(${PROJECT_NAME} PRIVATE USE_TENSORFLOW)
endif()

# Install configuration
install(TARGETS ${PROJECT_NAME}
    ARCHIVE DESTINATION lib
    LIBRARY DESTINATION lib
    RUNTIME DESTINATION bin)

install(FILES "config/detection_config.json"
    DESTINATION share/${PROJECT_NAME})
