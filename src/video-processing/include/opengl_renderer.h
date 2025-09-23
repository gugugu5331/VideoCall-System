#pragma once

#include "common.h"
#include "shader_manager.h"
#include "texture_manager.h"

namespace VideoProcessing {

class OpenGLRenderer {
public:
    OpenGLRenderer();
    ~OpenGLRenderer();

    // 初始化渲染器
    bool Initialize(int window_width, int window_height);
    
    // 释放资源
    void Release();
    
    // 窗口管理
    GLFWwindow* GetWindow() const { return window_; }
    void SetWindowSize(int width, int height);
    void GetWindowSize(int& width, int& height);
    bool ShouldClose() const;
    void SwapBuffers();
    void PollEvents();
    
    // 渲染控制
    void BeginFrame();
    void EndFrame();
    void Clear(const glm::vec4& color = glm::vec4(0.0f, 0.0f, 0.0f, 1.0f));
    
    // 基础渲染
    void RenderQuad();
    void RenderCube();
    void RenderSphere(int segments = 32);
    void RenderPlane(float width = 2.0f, float height = 2.0f);
    
    // 纹理渲染
    void RenderTexturedQuad(const std::string& texture_name, const glm::mat4& transform = glm::mat4(1.0f));
    void RenderVideoFrame(const cv::Mat& frame, const glm::mat4& transform = glm::mat4(1.0f));
    
    // 3D渲染
    void RenderMesh(const std::vector<float>& vertices, const std::vector<unsigned int>& indices,
                   const std::string& texture_name = "", const glm::mat4& model = glm::mat4(1.0f));
    
    // 面部贴图渲染
    void RenderFaceSticker(const cv::Mat& frame, const FaceInfo& face_info, 
                          const std::string& sticker_texture, float scale = 1.0f);
    void RenderFaceMask(const cv::Mat& frame, const FaceInfo& face_info, 
                       const std::string& mask_texture, float opacity = 0.8f);
    
    // 背景替换
    void RenderBackgroundReplacement(const cv::Mat& frame, const cv::Mat& mask, 
                                   const std::string& background_texture);
    
    // 粒子效果
    struct Particle {
        glm::vec3 position;
        glm::vec3 velocity;
        glm::vec4 color;
        float life;
        float size;
    };
    
    void InitializeParticleSystem(int max_particles = 1000);
    void UpdateParticles(float delta_time);
    void RenderParticles();
    void EmitParticles(const glm::vec3& position, int count, const glm::vec4& color);
    
    // 后处理效果
    void SetupFramebuffer(int width, int height);
    void BeginOffscreenRender();
    void EndOffscreenRender();
    void RenderFramebufferToScreen(const std::string& shader_name = "basic");
    
    // 多重采样抗锯齿
    void EnableMSAA(int samples = 4);
    void DisableMSAA();
    
    // 阴影映射
    void SetupShadowMapping(int shadow_map_size = 1024);
    void BeginShadowPass(const glm::vec3& light_position, const glm::vec3& light_target);
    void EndShadowPass();
    void RenderWithShadows(const glm::vec3& light_position);
    
    // 环境映射
    void SetupEnvironmentMapping(const std::string& cubemap_name);
    void RenderWithEnvironmentMapping(const glm::mat4& model, const std::string& texture_name = "");
    
    // 变形和动画
    void RenderMorphedMesh(const std::vector<float>& vertices1, const std::vector<float>& vertices2,
                          const std::vector<unsigned int>& indices, float morph_factor);
    
    // 实时变形
    void ApplyFaceDeformation(cv::Mat& frame, const FaceInfo& face_info, 
                             const std::vector<glm::vec2>& deformation_vectors);
    
    // 相机控制
    void SetViewMatrix(const glm::mat4& view) { view_matrix_ = view; }
    void SetProjectionMatrix(const glm::mat4& projection) { projection_matrix_ = projection; }
    void SetCameraPosition(const glm::vec3& position) { camera_position_ = position; }
    
    glm::mat4 GetViewMatrix() const { return view_matrix_; }
    glm::mat4 GetProjectionMatrix() const { return projection_matrix_; }
    glm::vec3 GetCameraPosition() const { return camera_position_; }
    
    // 光照设置
    struct Light {
        glm::vec3 position;
        glm::vec3 direction;
        glm::vec3 color;
        float intensity;
        float attenuation;
        int type; // 0=directional, 1=point, 2=spot
    };
    
    void AddLight(const Light& light);
    void RemoveLight(int index);
    void ClearLights();
    void UpdateLightUniforms(const std::string& shader_name);
    
    // 材质系统
    struct Material {
        glm::vec3 ambient;
        glm::vec3 diffuse;
        glm::vec3 specular;
        float shininess;
        std::string diffuse_texture;
        std::string normal_texture;
        std::string specular_texture;
    };
    
    void SetMaterial(const Material& material);
    void UpdateMaterialUniforms(const std::string& shader_name);
    
    // 渲染统计
    struct RenderStats {
        int draw_calls;
        int triangles_rendered;
        int vertices_processed;
        float frame_time;
        float gpu_time;
    };
    
    RenderStats GetRenderStats() const { return render_stats_; }
    void ResetRenderStats();
    
    // 调试功能
    void EnableWireframe(bool enable);
    void EnableDepthTest(bool enable);
    void EnableBlending(bool enable);
    void EnableCulling(bool enable);
    
    // 截图功能
    cv::Mat CaptureFramebuffer();
    bool SaveScreenshot(const std::string& filename);
    
    // 获取管理器
    ShaderManager& GetShaderManager() { return shader_manager_; }
    TextureManager& GetTextureManager() { return texture_manager_; }

private:
    GLFWwindow* window_;
    int window_width_;
    int window_height_;
    
    // 管理器
    ShaderManager shader_manager_;
    TextureManager texture_manager_;
    
    // 基础几何体VAO/VBO
    GLuint quad_VAO_, quad_VBO_;
    GLuint cube_VAO_, cube_VBO_;
    GLuint sphere_VAO_, sphere_VBO_, sphere_EBO_;
    
    // 帧缓冲
    GLuint framebuffer_;
    GLuint color_texture_;
    GLuint depth_texture_;
    GLuint rbo_;
    
    // MSAA
    GLuint msaa_framebuffer_;
    GLuint msaa_color_texture_;
    GLuint msaa_rbo_;
    int msaa_samples_;
    
    // 阴影映射
    GLuint shadow_framebuffer_;
    GLuint shadow_map_;
    int shadow_map_size_;
    glm::mat4 light_space_matrix_;
    
    // 粒子系统
    std::vector<Particle> particles_;
    GLuint particle_VAO_, particle_VBO_;
    int max_particles_;
    
    // 相机和变换
    glm::mat4 view_matrix_;
    glm::mat4 projection_matrix_;
    glm::vec3 camera_position_;
    
    // 光照和材质
    std::vector<Light> lights_;
    Material current_material_;
    
    // 渲染状态
    RenderStats render_stats_;
    bool wireframe_enabled_;
    bool depth_test_enabled_;
    bool blending_enabled_;
    bool culling_enabled_;
    
    // 初始化函数
    bool InitializeGLFW();
    bool InitializeGLEW();
    void InitializeGeometry();
    void InitializeFramebuffers();
    void InitializeParticles();
    
    // 几何体生成
    void GenerateQuad();
    void GenerateCube();
    void GenerateSphere(int segments);
    
    // 回调函数
    static void FramebufferSizeCallback(GLFWwindow* window, int width, int height);
    static void ErrorCallback(int error, const char* description);
    
    // 工具函数
    void UpdateRenderStats();
    void CheckFramebufferStatus();
    glm::mat4 CalculateLightSpaceMatrix(const glm::vec3& light_pos, const glm::vec3& light_target);
};

} // namespace VideoProcessing
