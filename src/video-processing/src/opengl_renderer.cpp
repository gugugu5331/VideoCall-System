#include "opengl_renderer.h"
#include "shader_manager.h"
#include <GL/glew.h>
#include <GLFW/glfw3.h>
#include <opencv2/opencv.hpp>
#include <iostream>

OpenGLRenderer::OpenGLRenderer()
    : initialized_(false)
    , window_(nullptr)
    , texture_id_(0)
    , vao_(0)
    , vbo_(0)
    , shader_manager_(nullptr)
{
}

OpenGLRenderer::~OpenGLRenderer() {
    cleanup();
}

bool OpenGLRenderer::initialize() {
    if (initialized_) {
        return true;
    }

    try {
        // 初始化GLFW
        if (!glfwInit()) {
            std::cerr << "Failed to initialize GLFW" << std::endl;
            return false;
        }

        // 设置OpenGL版本
        glfwWindowHint(GLFW_CONTEXT_VERSION_MAJOR, 3);
        glfwWindowHint(GLFW_CONTEXT_VERSION_MINOR, 3);
        glfwWindowHint(GLFW_OPENGL_PROFILE, GLFW_OPENGL_CORE_PROFILE);
        glfwWindowHint(GLFW_VISIBLE, GLFW_FALSE); // 隐藏窗口，仅用于离屏渲染

        // 创建窗口
        window_ = glfwCreateWindow(1280, 720, "Video Processing", nullptr, nullptr);
        if (!window_) {
            std::cerr << "Failed to create GLFW window" << std::endl;
            glfwTerminate();
            return false;
        }

        glfwMakeContextCurrent(window_);

        // 初始化GLEW
        if (glewInit() != GLEW_OK) {
            std::cerr << "Failed to initialize GLEW" << std::endl;
            return false;
        }

        std::cout << "OpenGL Version: " << glGetString(GL_VERSION) << std::endl;
        std::cout << "GLSL Version: " << glGetString(GL_SHADING_LANGUAGE_VERSION) << std::endl;

        // 初始化着色器管理器
        shader_manager_ = std::make_unique<ShaderManager>();
        if (!shader_manager_->initialize()) {
            std::cerr << "Failed to initialize ShaderManager" << std::endl;
            return false;
        }

        // 设置OpenGL状态
        setupOpenGL();

        // 创建渲染资源
        createRenderResources();

        initialized_ = true;
        std::cout << "OpenGLRenderer initialized successfully" << std::endl;
        return true;

    } catch (const std::exception& e) {
        std::cerr << "Error initializing OpenGLRenderer: " << e.what() << std::endl;
        return false;
    }
}

void OpenGLRenderer::cleanup() {
    if (vao_ != 0) {
        glDeleteVertexArrays(1, &vao_);
        vao_ = 0;
    }

    if (vbo_ != 0) {
        glDeleteBuffers(1, &vbo_);
        vbo_ = 0;
    }

    if (texture_id_ != 0) {
        glDeleteTextures(1, &texture_id_);
        texture_id_ = 0;
    }

    if (shader_manager_) {
        shader_manager_->cleanup();
        shader_manager_.reset();
    }

    if (window_) {
        glfwDestroyWindow(window_);
        window_ = nullptr;
    }

    glfwTerminate();
    initialized_ = false;
}

void OpenGLRenderer::renderFrame(cv::Mat& frame) {
    if (!initialized_ || frame.empty()) {
        return;
    }

    try {
        glfwMakeContextCurrent(window_);

        // 设置视口
        glViewport(0, 0, frame.cols, frame.rows);

        // 清除缓冲区
        glClear(GL_COLOR_BUFFER_BIT);

        // 上传纹理
        uploadTexture(frame);

        // 使用基础着色器
        GLuint shader_program = shader_manager_->getShaderProgram("basic");
        if (shader_program != 0) {
            glUseProgram(shader_program);

            // 绑定纹理
            glActiveTexture(GL_TEXTURE0);
            glBindTexture(GL_TEXTURE_2D, texture_id_);
            glUniform1i(glGetUniformLocation(shader_program, "u_texture"), 0);

            // 设置时间uniform（用于动画效果）
            float time = static_cast<float>(glfwGetTime());
            glUniform1f(glGetUniformLocation(shader_program, "u_time"), time);

            // 设置分辨率uniform
            glUniform2f(glGetUniformLocation(shader_program, "u_resolution"), 
                       static_cast<float>(frame.cols), static_cast<float>(frame.rows));

            // 渲染四边形
            glBindVertexArray(vao_);
            glDrawArrays(GL_TRIANGLES, 0, 6);
            glBindVertexArray(0);

            glUseProgram(0);
        }

        // 读取渲染结果回到CPU
        readFramebuffer(frame);

        // 交换缓冲区
        glfwSwapBuffers(window_);

    } catch (const std::exception& e) {
        std::cerr << "Error rendering frame: " << e.what() << std::endl;
    }
}

void OpenGLRenderer::applyShaderEffect(cv::Mat& frame, const std::string& shader_name) {
    if (!initialized_ || frame.empty()) {
        return;
    }

    try {
        glfwMakeContextCurrent(window_);

        // 设置视口
        glViewport(0, 0, frame.cols, frame.rows);

        // 清除缓冲区
        glClear(GL_COLOR_BUFFER_BIT);

        // 上传纹理
        uploadTexture(frame);

        // 使用指定的着色器
        GLuint shader_program = shader_manager_->getShaderProgram(shader_name);
        if (shader_program != 0) {
            glUseProgram(shader_program);

            // 绑定纹理
            glActiveTexture(GL_TEXTURE0);
            glBindTexture(GL_TEXTURE_2D, texture_id_);
            glUniform1i(glGetUniformLocation(shader_program, "u_texture"), 0);

            // 设置通用uniform
            float time = static_cast<float>(glfwGetTime());
            glUniform1f(glGetUniformLocation(shader_program, "u_time"), time);
            glUniform2f(glGetUniformLocation(shader_program, "u_resolution"), 
                       static_cast<float>(frame.cols), static_cast<float>(frame.rows));

            // 渲染四边形
            glBindVertexArray(vao_);
            glDrawArrays(GL_TRIANGLES, 0, 6);
            glBindVertexArray(0);

            glUseProgram(0);
        }

        // 读取渲染结果
        readFramebuffer(frame);

        glfwSwapBuffers(window_);

    } catch (const std::exception& e) {
        std::cerr << "Error applying shader effect: " << e.what() << std::endl;
    }
}

void OpenGLRenderer::setupOpenGL() {
    // 启用混合
    glEnable(GL_BLEND);
    glBlendFunc(GL_SRC_ALPHA, GL_ONE_MINUS_SRC_ALPHA);

    // 设置清除颜色
    glClearColor(0.0f, 0.0f, 0.0f, 1.0f);

    // 禁用深度测试（2D渲染）
    glDisable(GL_DEPTH_TEST);
}

void OpenGLRenderer::createRenderResources() {
    // 创建全屏四边形的顶点数据
    float vertices[] = {
        // 位置        // 纹理坐标
        -1.0f, -1.0f,  0.0f, 0.0f,
         1.0f, -1.0f,  1.0f, 0.0f,
         1.0f,  1.0f,  1.0f, 1.0f,

        -1.0f, -1.0f,  0.0f, 0.0f,
         1.0f,  1.0f,  1.0f, 1.0f,
        -1.0f,  1.0f,  0.0f, 1.0f
    };

    // 创建VAO和VBO
    glGenVertexArrays(1, &vao_);
    glGenBuffers(1, &vbo_);

    glBindVertexArray(vao_);
    glBindBuffer(GL_ARRAY_BUFFER, vbo_);
    glBufferData(GL_ARRAY_BUFFER, sizeof(vertices), vertices, GL_STATIC_DRAW);

    // 位置属性
    glVertexAttribPointer(0, 2, GL_FLOAT, GL_FALSE, 4 * sizeof(float), (void*)0);
    glEnableVertexAttribArray(0);

    // 纹理坐标属性
    glVertexAttribPointer(1, 2, GL_FLOAT, GL_FALSE, 4 * sizeof(float), (void*)(2 * sizeof(float)));
    glEnableVertexAttribArray(1);

    glBindVertexArray(0);

    // 创建纹理
    glGenTextures(1, &texture_id_);
    glBindTexture(GL_TEXTURE_2D, texture_id_);
    glTexParameteri(GL_TEXTURE_2D, GL_TEXTURE_MIN_FILTER, GL_LINEAR);
    glTexParameteri(GL_TEXTURE_2D, GL_TEXTURE_MAG_FILTER, GL_LINEAR);
    glTexParameteri(GL_TEXTURE_2D, GL_TEXTURE_WRAP_S, GL_CLAMP_TO_EDGE);
    glTexParameteri(GL_TEXTURE_2D, GL_TEXTURE_WRAP_T, GL_CLAMP_TO_EDGE);
    glBindTexture(GL_TEXTURE_2D, 0);
}

void OpenGLRenderer::uploadTexture(const cv::Mat& frame) {
    if (frame.empty()) {
        return;
    }

    // 确保图像格式正确
    cv::Mat rgb_frame;
    if (frame.channels() == 3) {
        cv::cvtColor(frame, rgb_frame, cv::COLOR_BGR2RGB);
    } else if (frame.channels() == 4) {
        cv::cvtColor(frame, rgb_frame, cv::COLOR_BGRA2RGBA);
    } else {
        rgb_frame = frame;
    }

    // 上传到GPU
    glBindTexture(GL_TEXTURE_2D, texture_id_);
    
    GLenum format = (rgb_frame.channels() == 4) ? GL_RGBA : GL_RGB;
    glTexImage2D(GL_TEXTURE_2D, 0, format, rgb_frame.cols, rgb_frame.rows, 
                 0, format, GL_UNSIGNED_BYTE, rgb_frame.data);
    
    glBindTexture(GL_TEXTURE_2D, 0);
}

void OpenGLRenderer::readFramebuffer(cv::Mat& frame) {
    if (frame.empty()) {
        return;
    }

    // 读取帧缓冲区数据
    cv::Mat result(frame.rows, frame.cols, CV_8UC3);
    glReadPixels(0, 0, frame.cols, frame.rows, GL_RGB, GL_UNSIGNED_BYTE, result.data);

    // OpenGL的Y轴是反向的，需要翻转
    cv::flip(result, result, 0);

    // 转换回BGR格式
    cv::cvtColor(result, frame, cv::COLOR_RGB2BGR);
}

void OpenGLRenderer::setShaderUniform(const std::string& shader_name, const std::string& uniform_name, float value) {
    if (!shader_manager_) {
        return;
    }

    GLuint program = shader_manager_->getShaderProgram(shader_name);
    if (program != 0) {
        glUseProgram(program);
        GLint location = glGetUniformLocation(program, uniform_name.c_str());
        if (location != -1) {
            glUniform1f(location, value);
        }
        glUseProgram(0);
    }
}

void OpenGLRenderer::setShaderUniform(const std::string& shader_name, const std::string& uniform_name, const cv::Vec2f& value) {
    if (!shader_manager_) {
        return;
    }

    GLuint program = shader_manager_->getShaderProgram(shader_name);
    if (program != 0) {
        glUseProgram(program);
        GLint location = glGetUniformLocation(program, uniform_name.c_str());
        if (location != -1) {
            glUniform2f(location, value[0], value[1]);
        }
        glUseProgram(0);
    }
}

void OpenGLRenderer::setShaderUniform(const std::string& shader_name, const std::string& uniform_name, const cv::Vec3f& value) {
    if (!shader_manager_) {
        return;
    }

    GLuint program = shader_manager_->getShaderProgram(shader_name);
    if (program != 0) {
        glUseProgram(program);
        GLint location = glGetUniformLocation(program, uniform_name.c_str());
        if (location != -1) {
            glUniform3f(location, value[0], value[1], value[2]);
        }
        glUseProgram(0);
    }
}

std::vector<std::string> OpenGLRenderer::getAvailableShaders() const {
    if (shader_manager_) {
        return shader_manager_->getAvailableShaders();
    }
    return {};
}

bool OpenGLRenderer::loadShader(const std::string& name, const std::string& vertex_path, const std::string& fragment_path) {
    if (shader_manager_) {
        return shader_manager_->loadShader(name, vertex_path, fragment_path);
    }
    return false;
}
