cmake_minimum_required(VERSION 3.10)
set(CMAKE_CXX_STANDARD 17)
set(CMAKE_CXX_STANDARD_REQUIRED ON)
set(INSTALL_ROOT_DIR "${CMAKE_SOURCE_DIR}/../install")
set(CMAKE_INSTALL_PREFIX "${INSTALL_ROOT_DIR}" CACHE PATH "Install path" FORCE)

include_directories("../hybrid-comm/include/")
include_directories("../hybrid-comm/include/libzmq")
include_directories("./include/")
include_directories("../utils/")

file(GLOB SRCS "src/*.cpp" "../hybrid-comm/src/*.cpp")

add_library(stackflow STATIC ${SRCS})


find_package(eventpp REQUIRED)
target_link_libraries(stackflow
    eventpp
    utilities
    zmq
    # simdjson_component
)

install(TARGETS stackflow
    ARCHIVE DESTINATION lib
    LIBRARY DESTINATION lib
)


