#pragma once

#include "common.h"

namespace VideoProcessing {

class ShaderManager {
public:
    ShaderManager();
    ~ShaderManager();

    // 初始化着色器管理器
    bool Initialize();
    
    // 释放资源
    void Release();
    
    // 加载着色器程序
    bool LoadShader(const std::string& name, const std::string& vertex_path, 
                   const std::string& fragment_path);
    
    // 使用着色器程序
    bool UseShader(const std::string& name);
    
    // 获取着色器程序ID
    GLuint GetShaderProgram(const std::string& name);
    
    // 设置uniform变量
    void SetUniform(const std::string& shader_name, const std::string& uniform_name, int value);
    void SetUniform(const std::string& shader_name, const std::string& uniform_name, float value);
    void SetUniform(const std::string& shader_name, const std::string& uniform_name, const glm::vec2& value);
    void SetUniform(const std::string& shader_name, const std::string& uniform_name, const glm::vec3& value);
    void SetUniform(const std::string& shader_name, const std::string& uniform_name, const glm::vec4& value);
    void SetUniform(const std::string& shader_name, const std::string& uniform_name, const glm::mat4& value);
    
    // 获取uniform位置
    GLint GetUniformLocation(const std::string& shader_name, const std::string& uniform_name);
    
    // 重新加载着色器（用于调试）
    bool ReloadShader(const std::string& name);
    
    // 获取所有着色器名称
    std::vector<std::string> GetShaderNames() const;
    
    // 检查着色器是否存在
    bool HasShader(const std::string& name) const;
    
    // 创建默认着色器
    void CreateDefaultShaders();

private:
    std::map<std::string, ShaderInfo> shaders_;
    std::string current_shader_;
    
    // 编译着色器
    GLuint CompileShader(const std::string& source, GLenum shader_type);
    
    // 链接着色器程序
    GLuint LinkProgram(GLuint vertex_shader, GLuint fragment_shader);
    
    // 读取着色器文件
    std::string ReadShaderFile(const std::string& file_path);
    
    // 检查编译错误
    bool CheckCompileErrors(GLuint shader, const std::string& type);
    
    // 缓存uniform位置
    void CacheUniformLocations(const std::string& shader_name);
    
    // 验证着色器程序
    bool ValidateProgram(GLuint program);
};

} // namespace VideoProcessing
