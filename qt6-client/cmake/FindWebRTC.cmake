# FindWebRTC.cmake - 查找和配置WebRTC库
#
# 使用方法:
#   find_package(WebRTC REQUIRED)
#
# 定义的变量:
#   WebRTC_FOUND - 是否找到WebRTC
#   WebRTC_INCLUDE_DIRS - WebRTC头文件目录
#   WebRTC_LIBRARIES - WebRTC库文件
#   WebRTC_VERSION - WebRTC版本

# 设置WebRTC搜索路径
set(WebRTC_ROOT_HINTS
    ${WebRTC_ROOT}
    $ENV{WEBRTC_ROOT}
    ${CMAKE_SOURCE_DIR}/third_party/webrtc
    ${CMAKE_SOURCE_DIR}/../webrtc
)

# 根据平台设置库名称
if(WIN32)
    set(WebRTC_LIB_NAME "webrtc.lib")
    set(WebRTC_PLATFORM "win")
elseif(APPLE)
    set(WebRTC_LIB_NAME "libwebrtc.a")
    set(WebRTC_PLATFORM "mac")
else()
    set(WebRTC_LIB_NAME "libwebrtc.a")
    set(WebRTC_PLATFORM "linux")
endif()

# 查找头文件
find_path(WebRTC_INCLUDE_DIR
    NAMES
        api/peer_connection_interface.h
        api/create_peerconnection_factory.h
    PATHS
        ${WebRTC_ROOT_HINTS}
    PATH_SUFFIXES
        include
        src
)

# 查找库文件
find_library(WebRTC_LIBRARY
    NAMES
        webrtc
        ${WebRTC_LIB_NAME}
    PATHS
        ${WebRTC_ROOT_HINTS}
    PATH_SUFFIXES
        lib
        lib/${WebRTC_PLATFORM}
        out/Release/obj
        out/Debug/obj
)

# 处理标准参数
include(FindPackageHandleStandardArgs)
find_package_handle_standard_args(WebRTC
    REQUIRED_VARS
        WebRTC_LIBRARY
        WebRTC_INCLUDE_DIR
    VERSION_VAR
        WebRTC_VERSION
)

if(WebRTC_FOUND)
    set(WebRTC_LIBRARIES ${WebRTC_LIBRARY})
    set(WebRTC_INCLUDE_DIRS ${WebRTC_INCLUDE_DIR})
    
    # 创建导入目标
    if(NOT TARGET WebRTC::WebRTC)
        add_library(WebRTC::WebRTC UNKNOWN IMPORTED)
        set_target_properties(WebRTC::WebRTC PROPERTIES
            IMPORTED_LOCATION "${WebRTC_LIBRARY}"
            INTERFACE_INCLUDE_DIRECTORIES "${WebRTC_INCLUDE_DIR}"
        )
        
        # 添加WebRTC依赖的系统库
        if(WIN32)
            target_link_libraries(WebRTC::WebRTC INTERFACE
                winmm.lib
                dmoguids.lib
                wmcodecdspuuid.lib
                msdmo.lib
                strmiids.lib
                secur32.lib
                iphlpapi.lib
            )
        elseif(APPLE)
            target_link_libraries(WebRTC::WebRTC INTERFACE
                "-framework Foundation"
                "-framework AVFoundation"
                "-framework CoreAudio"
                "-framework CoreMedia"
                "-framework CoreVideo"
                "-framework AudioToolbox"
                "-framework VideoToolbox"
            )
        else()
            target_link_libraries(WebRTC::WebRTC INTERFACE
                dl
                pthread
                X11
                asound
            )
        endif()
        
        # 添加编译定义
        target_compile_definitions(WebRTC::WebRTC INTERFACE
            WEBRTC_POSIX
            WEBRTC_LINUX  # 或 WEBRTC_WIN, WEBRTC_MAC
        )
    endif()
    
    message(STATUS "Found WebRTC: ${WebRTC_LIBRARY}")
    message(STATUS "WebRTC include dir: ${WebRTC_INCLUDE_DIR}")
endif()

mark_as_advanced(
    WebRTC_INCLUDE_DIR
    WebRTC_LIBRARY
)

