cmake_minimum_required(VERSION 3.12)
project(unit_manager VERSION 1.6)

set(CMAKE_CXX_STANDARD 17)
set(CMAKE_CXX_STANDARD_REQUIRED ON)
set(CMAKE_BUILD_TYPE Debug)  
set(REQUIREMENTS  
    pthread 
    util 
    glog
    )

include_directories(
    ${CMAKE_SOURCE_DIR}/include
    ../hybrid-comm/include
    ../infra-controller/include
    ../utils
    ../network/include
    "/usr/local/include/"
    "include" "." 
)

file(GLOB_RECURSE SRCS "src/main.cpp" 
"src/remote_server.cpp" 
"src/config.cpp" 
"src/tcp_comm.cpp"
"src/remote_action.cpp" 
"src/unit_data.cpp" 
"src/zmq_bus.cpp"  
"${CMAKE_SOURCE_DIR}/../network/src/*.cc")



# find_package(simdjson REQUIRED)
# link_directories(/usr/lib/x86_64-linux-gnu)

add_executable(${PROJECT_NAME}  ${SRCS})

target_link_libraries(${PROJECT_NAME} PRIVATE 
    pthread 
    util 
    glog
    pgm 
    sodium 
    norm
    bsd
    gssapi_krb5
    zmq
    /usr/local/lib/libsimdjson.a 
    ${CMAKE_SOURCE_DIR}/../install/lib/libstackflow.a)



install(TARGETS ${PROJECT_NAME}
    ARCHIVE DESTINATION lib
    LIBRARY DESTINATION lib
    RUNTIME DESTINATION bin)

install(FILES "sys_config.json"
    DESTINATION share/${PROJECT_NAME})
