#pragma once

#include "common.h"

namespace VideoProcessing {

class TextureManager {
public:
    TextureManager();
    ~TextureManager();

    // 初始化纹理管理器
    bool Initialize();
    
    // 释放资源
    void Release();
    
    // 从文件加载纹理
    bool LoadTexture(const std::string& name, const std::string& file_path, bool flip_vertically = true);
    
    // 从OpenCV Mat创建纹理
    bool CreateTextureFromMat(const std::string& name, const cv::Mat& image, bool flip_vertically = false);
    
    // 创建空纹理
    bool CreateEmptyTexture(const std::string& name, int width, int height, GLenum format = GL_RGB);
    
    // 更新纹理数据
    bool UpdateTexture(const std::string& name, const cv::Mat& image);
    bool UpdateTextureRegion(const std::string& name, const cv::Mat& image, int x, int y);
    
    // 获取纹理
    GLuint GetTexture(const std::string& name);
    TextureInfo GetTextureInfo(const std::string& name);
    
    // 绑定纹理
    void BindTexture(const std::string& name, int texture_unit = 0);
    void UnbindTexture(int texture_unit = 0);
    
    // 删除纹理
    void DeleteTexture(const std::string& name);
    
    // 纹理操作
    cv::Mat ReadTextureToMat(const std::string& name);
    bool SaveTextureToFile(const std::string& name, const std::string& file_path);
    
    // 纹理效果
    bool ApplyTextureFilter(const std::string& name, FilterType filter_type, const EffectParams& params);
    bool BlendTextures(const std::string& result_name, const std::string& texture1, 
                      const std::string& texture2, float blend_factor);
    
    // 动态纹理（用于视频流）
    bool CreateVideoTexture(const std::string& name, int width, int height);
    bool UpdateVideoTexture(const std::string& name, const cv::Mat& frame);
    
    // 立方体贴图
    bool LoadCubemap(const std::string& name, const std::vector<std::string>& face_paths);
    bool CreateCubemapFromMats(const std::string& name, const std::vector<cv::Mat>& faces);
    
    // 纹理数组
    bool CreateTextureArray(const std::string& name, const std::vector<cv::Mat>& textures);
    bool UpdateTextureArrayLayer(const std::string& name, int layer, const cv::Mat& texture);
    
    // 帧缓冲纹理
    bool CreateFramebufferTexture(const std::string& name, int width, int height, 
                                 GLenum internal_format = GL_RGB8, GLenum format = GL_RGB, 
                                 GLenum type = GL_UNSIGNED_BYTE);
    
    // 深度纹理
    bool CreateDepthTexture(const std::string& name, int width, int height);
    
    // 纹理压缩
    bool CompressTexture(const std::string& name, GLenum compression_format);
    
    // 纹理生成
    bool GenerateCheckerboardTexture(const std::string& name, int width, int height, 
                                    int checker_size, const glm::vec3& color1, const glm::vec3& color2);
    bool GenerateNoiseTexture(const std::string& name, int width, int height, float frequency = 1.0f);
    bool GenerateGradientTexture(const std::string& name, int width, int height, 
                                const glm::vec3& start_color, const glm::vec3& end_color, bool horizontal = true);
    
    // 纹理信息
    std::vector<std::string> GetTextureNames() const;
    bool HasTexture(const std::string& name) const;
    size_t GetTextureCount() const { return textures_.size(); }
    size_t GetTotalMemoryUsage() const;
    
    // 纹理设置
    void SetTextureFiltering(const std::string& name, GLenum min_filter, GLenum mag_filter);
    void SetTextureWrapping(const std::string& name, GLenum wrap_s, GLenum wrap_t);
    void SetTextureAnisotropy(const std::string& name, float anisotropy);
    void GenerateMipmaps(const std::string& name);
    
    // 批量操作
    void DeleteAllTextures();
    void ReloadAllTextures();
    
    // 调试功能
    void PrintTextureInfo(const std::string& name);
    void PrintAllTexturesInfo();
    bool ValidateTexture(const std::string& name);

private:
    std::map<std::string, TextureInfo> textures_;
    std::map<std::string, std::string> texture_paths_; // 用于重新加载
    
    // OpenGL状态
    int max_texture_units_;
    int max_texture_size_;
    float max_anisotropy_;
    
    // 内部辅助函数
    GLenum DetermineFormat(const cv::Mat& image);
    GLenum DetermineInternalFormat(const cv::Mat& image);
    GLenum DetermineType(const cv::Mat& image);
    
    // 图像处理
    cv::Mat PrepareImageForTexture(const cv::Mat& image, bool flip_vertically);
    cv::Mat ConvertToRGB(const cv::Mat& image);
    
    // 纹理创建辅助
    GLuint CreateGLTexture(const cv::Mat& image, GLenum target = GL_TEXTURE_2D);
    bool SetTextureParameters(GLuint texture_id, GLenum target = GL_TEXTURE_2D);
    
    // 错误检查
    bool CheckTextureComplete(GLuint texture_id, GLenum target = GL_TEXTURE_2D);
    void LogTextureError(const std::string& operation, const std::string& texture_name);
    
    // 内存管理
    void UpdateMemoryUsage(const std::string& name, size_t size);
    size_t CalculateTextureSize(int width, int height, GLenum format, GLenum type);
    
    // 性能优化
    void OptimizeTextureStorage();
    bool IsTextureResident(const std::string& name);
};

} // namespace VideoProcessing
