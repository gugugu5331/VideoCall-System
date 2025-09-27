#include "shader_manager.h"
#include <GL/glew.h>
#include <iostream>
#include <fstream>
#include <sstream>

ShaderManager::ShaderManager()
    : initialized_(false)
{
}

ShaderManager::~ShaderManager() {
    cleanup();
}

bool ShaderManager::initialize() {
    if (initialized_) {
        return true;
    }

    try {
        // 加载基础着色器
        loadDefaultShaders();
        
        initialized_ = true;
        std::cout << "ShaderManager initialized successfully" << std::endl;
        return true;
        
    } catch (const std::exception& e) {
        std::cerr << "Error initializing ShaderManager: " << e.what() << std::endl;
        return false;
    }
}

void ShaderManager::cleanup() {
    for (auto& shader_pair : shader_programs_) {
        glDeleteProgram(shader_pair.second);
    }
    shader_programs_.clear();
    initialized_ = false;
}

bool ShaderManager::loadShader(const std::string& name, const std::string& vertex_path, const std::string& fragment_path) {
    try {
        // 读取着色器源码
        std::string vertex_source = readShaderFile(vertex_path);
        std::string fragment_source = readShaderFile(fragment_path);

        if (vertex_source.empty() || fragment_source.empty()) {
            std::cerr << "Failed to read shader files: " << vertex_path << ", " << fragment_path << std::endl;
            return false;
        }

        // 编译着色器
        GLuint vertex_shader = compileShader(vertex_source, GL_VERTEX_SHADER);
        GLuint fragment_shader = compileShader(fragment_source, GL_FRAGMENT_SHADER);

        if (vertex_shader == 0 || fragment_shader == 0) {
            if (vertex_shader != 0) glDeleteShader(vertex_shader);
            if (fragment_shader != 0) glDeleteShader(fragment_shader);
            return false;
        }

        // 创建着色器程序
        GLuint program = createShaderProgram(vertex_shader, fragment_shader);
        
        // 清理着色器对象
        glDeleteShader(vertex_shader);
        glDeleteShader(fragment_shader);

        if (program == 0) {
            return false;
        }

        // 删除旧程序（如果存在）
        auto it = shader_programs_.find(name);
        if (it != shader_programs_.end()) {
            glDeleteProgram(it->second);
        }

        shader_programs_[name] = program;
        std::cout << "Loaded shader: " << name << std::endl;
        return true;

    } catch (const std::exception& e) {
        std::cerr << "Error loading shader " << name << ": " << e.what() << std::endl;
        return false;
    }
}

GLuint ShaderManager::getShaderProgram(const std::string& name) const {
    auto it = shader_programs_.find(name);
    if (it != shader_programs_.end()) {
        return it->second;
    }
    return 0;
}

std::vector<std::string> ShaderManager::getAvailableShaders() const {
    std::vector<std::string> shader_names;
    for (const auto& shader_pair : shader_programs_) {
        shader_names.push_back(shader_pair.first);
    }
    return shader_names;
}

std::string ShaderManager::readShaderFile(const std::string& file_path) {
    std::ifstream file(file_path);
    if (!file.is_open()) {
        std::cerr << "Failed to open shader file: " << file_path << std::endl;
        return "";
    }

    std::stringstream buffer;
    buffer << file.rdbuf();
    return buffer.str();
}

GLuint ShaderManager::compileShader(const std::string& source, GLenum shader_type) {
    GLuint shader = glCreateShader(shader_type);
    const char* source_cstr = source.c_str();
    glShaderSource(shader, 1, &source_cstr, nullptr);
    glCompileShader(shader);

    // 检查编译状态
    GLint success;
    glGetShaderiv(shader, GL_COMPILE_STATUS, &success);
    if (!success) {
        GLchar info_log[512];
        glGetShaderInfoLog(shader, 512, nullptr, info_log);
        std::cerr << "Shader compilation failed: " << info_log << std::endl;
        glDeleteShader(shader);
        return 0;
    }

    return shader;
}

GLuint ShaderManager::createShaderProgram(GLuint vertex_shader, GLuint fragment_shader) {
    GLuint program = glCreateProgram();
    glAttachShader(program, vertex_shader);
    glAttachShader(program, fragment_shader);
    glLinkProgram(program);

    // 检查链接状态
    GLint success;
    glGetProgramiv(program, GL_LINK_STATUS, &success);
    if (!success) {
        GLchar info_log[512];
        glGetProgramInfoLog(program, 512, nullptr, info_log);
        std::cerr << "Shader program linking failed: " << info_log << std::endl;
        glDeleteProgram(program);
        return 0;
    }

    return program;
}

void ShaderManager::loadDefaultShaders() {
    // 基础顶点着色器源码
    std::string basic_vertex_source = R"(
#version 330 core
layout (location = 0) in vec2 aPos;
layout (location = 1) in vec2 aTexCoord;

out vec2 TexCoord;

void main()
{
    gl_Position = vec4(aPos, 0.0, 1.0);
    TexCoord = aTexCoord;
}
)";

    // 基础片段着色器源码
    std::string basic_fragment_source = R"(
#version 330 core
out vec4 FragColor;

in vec2 TexCoord;

uniform sampler2D u_texture;
uniform float u_time;
uniform vec2 u_resolution;

void main()
{
    FragColor = texture(u_texture, TexCoord);
}
)";

    // 模糊效果着色器
    std::string blur_fragment_source = R"(
#version 330 core
out vec4 FragColor;

in vec2 TexCoord;

uniform sampler2D u_texture;
uniform float u_time;
uniform vec2 u_resolution;

void main()
{
    vec2 texelSize = 1.0 / u_resolution;
    vec4 result = vec4(0.0);
    
    // 简单的高斯模糊
    for(int x = -2; x <= 2; x++) {
        for(int y = -2; y <= 2; y++) {
            vec2 offset = vec2(float(x), float(y)) * texelSize;
            result += texture(u_texture, TexCoord + offset);
        }
    }
    
    FragColor = result / 25.0;
}
)";

    // 边缘检测着色器
    std::string edge_fragment_source = R"(
#version 330 core
out vec4 FragColor;

in vec2 TexCoord;

uniform sampler2D u_texture;
uniform float u_time;
uniform vec2 u_resolution;

void main()
{
    vec2 texelSize = 1.0 / u_resolution;
    
    // Sobel边缘检测
    vec3 tl = texture(u_texture, TexCoord + vec2(-texelSize.x, -texelSize.y)).rgb;
    vec3 tm = texture(u_texture, TexCoord + vec2(0.0, -texelSize.y)).rgb;
    vec3 tr = texture(u_texture, TexCoord + vec2(texelSize.x, -texelSize.y)).rgb;
    vec3 ml = texture(u_texture, TexCoord + vec2(-texelSize.x, 0.0)).rgb;
    vec3 mm = texture(u_texture, TexCoord).rgb;
    vec3 mr = texture(u_texture, TexCoord + vec2(texelSize.x, 0.0)).rgb;
    vec3 bl = texture(u_texture, TexCoord + vec2(-texelSize.x, texelSize.y)).rgb;
    vec3 bm = texture(u_texture, TexCoord + vec2(0.0, texelSize.y)).rgb;
    vec3 br = texture(u_texture, TexCoord + vec2(texelSize.x, texelSize.y)).rgb;
    
    vec3 gx = -tl + tr - 2.0*ml + 2.0*mr - bl + br;
    vec3 gy = -tl - 2.0*tm - tr + bl + 2.0*bm + br;
    
    float edge = length(gx) + length(gy);
    FragColor = vec4(vec3(edge), 1.0);
}
)";

    // 复古效果着色器
    std::string vintage_fragment_source = R"(
#version 330 core
out vec4 FragColor;

in vec2 TexCoord;

uniform sampler2D u_texture;
uniform float u_time;
uniform vec2 u_resolution;

void main()
{
    vec4 color = texture(u_texture, TexCoord);
    
    // 棕褐色调
    float gray = dot(color.rgb, vec3(0.299, 0.587, 0.114));
    vec3 sepia = vec3(gray) * vec3(1.2, 1.0, 0.8);
    
    // 添加噪点
    float noise = fract(sin(dot(TexCoord * u_time, vec2(12.9898, 78.233))) * 43758.5453);
    sepia += (noise - 0.5) * 0.1;
    
    // 暗角效果
    vec2 center = TexCoord - 0.5;
    float vignette = 1.0 - dot(center, center) * 0.8;
    
    FragColor = vec4(sepia * vignette, color.a);
}
)";

    // 编译并创建着色器程序
    createShaderFromSource("basic", basic_vertex_source, basic_fragment_source);
    createShaderFromSource("blur", basic_vertex_source, blur_fragment_source);
    createShaderFromSource("edge", basic_vertex_source, edge_fragment_source);
    createShaderFromSource("vintage", basic_vertex_source, vintage_fragment_source);
}

bool ShaderManager::createShaderFromSource(const std::string& name, const std::string& vertex_source, const std::string& fragment_source) {
    try {
        // 编译着色器
        GLuint vertex_shader = compileShader(vertex_source, GL_VERTEX_SHADER);
        GLuint fragment_shader = compileShader(fragment_source, GL_FRAGMENT_SHADER);

        if (vertex_shader == 0 || fragment_shader == 0) {
            if (vertex_shader != 0) glDeleteShader(vertex_shader);
            if (fragment_shader != 0) glDeleteShader(fragment_shader);
            return false;
        }

        // 创建着色器程序
        GLuint program = createShaderProgram(vertex_shader, fragment_shader);
        
        // 清理着色器对象
        glDeleteShader(vertex_shader);
        glDeleteShader(fragment_shader);

        if (program == 0) {
            return false;
        }

        shader_programs_[name] = program;
        std::cout << "Created shader from source: " << name << std::endl;
        return true;

    } catch (const std::exception& e) {
        std::cerr << "Error creating shader from source " << name << ": " << e.what() << std::endl;
        return false;
    }
}
