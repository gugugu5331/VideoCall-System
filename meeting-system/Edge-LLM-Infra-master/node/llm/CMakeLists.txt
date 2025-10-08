cmake_minimum_required(VERSION 3.12)
project(llm)

set(CMAKE_CXX_STANDARD 17)
set(CMAKE_CXX_STANDARD_REQUIRED ON)
set(CMAKE_BUILD_TYPE Debug)
set(CMAKE_CXX_FLAGS_DEBUG "${CMAKE_CXX_FLAGS_DEBUG} -g -O0")

find_package(Threads REQUIRED)

include_directories(
    ${CMAKE_SOURCE_DIR}/include
    ../../hybrid-comm/include
    ../../infra-controller/include
    ../../utils
    /usr/local/include
)

file(GLOB SRCS
    src/main.cpp
    src/asr_task.cpp
    src/emotion_task.cpp
    src/synthesis_task.cpp
    src/whisper_asr_task.cpp
)

add_executable(${PROJECT_NAME} ${SRCS})

target_link_libraries(${PROJECT_NAME}
    PRIVATE
    pthread
    ${CMAKE_SOURCE_DIR}/../../install/lib/libstackflow.a
    zmq
    onnxruntime
)

