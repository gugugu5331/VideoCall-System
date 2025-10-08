#ifndef GL_VIDEO_RENDERER_H
#define GL_VIDEO_RENDERER_H

#include <QObject>
#include <QOpenGLWidget>
#include <QOpenGLFunctions>
#include <QOpenGLShaderProgram>
#include <QOpenGLTexture>
#include <QOpenGLBuffer>
#include <QOpenGLVertexArrayObject>
#include <QVideoFrame>
#include <QImage>
#include <memory>

/**
 * @brief GLVideoRenderer类 - OpenGL视频渲染器
 * 
 * 使用OpenGL进行GPU加速的视频渲染和效果处理
 */
class GLVideoRenderer : public QOpenGLWidget, protected QOpenGLFunctions
{
    Q_OBJECT
    Q_PROPERTY(bool effectsEnabled READ effectsEnabled WRITE setEffectsEnabled NOTIFY effectsEnabledChanged)
    Q_PROPERTY(float brightness READ brightness WRITE setBrightness NOTIFY brightnessChanged)
    Q_PROPERTY(float contrast READ contrast WRITE setContrast NOTIFY contrastChanged)
    Q_PROPERTY(float saturation READ saturation WRITE setSaturation NOTIFY saturationChanged)

public:
    explicit GLVideoRenderer(QWidget *parent = nullptr);
    ~GLVideoRenderer();

    // Properties
    bool effectsEnabled() const { return m_effectsEnabled; }
    float brightness() const { return m_brightness; }
    float contrast() const { return m_contrast; }
    float saturation() const { return m_saturation; }

    // Setters
    void setEffectsEnabled(bool enabled);
    void setBrightness(float value);  // -1.0 to 1.0
    void setContrast(float value);    // 0.0 to 2.0
    void setSaturation(float value);  // 0.0 to 2.0

    // Frame rendering
    Q_INVOKABLE void setVideoFrame(const QVideoFrame &frame);
    Q_INVOKABLE void setImage(const QImage &image);
    Q_INVOKABLE void clear();

signals:
    void effectsEnabledChanged();
    void brightnessChanged();
    void contrastChanged();
    void saturationChanged();
    void renderError(const QString &error);

protected:
    // OpenGL functions
    void initializeGL() override;
    void resizeGL(int w, int h) override;
    void paintGL() override;

private:
    // Shader initialization
    bool initShaders();
    bool loadShader(QOpenGLShaderProgram *program, const QString &vertexPath, const QString &fragmentPath);
    
    // Texture management
    void updateTexture(const QImage &image);
    void deleteTexture();

    // Rendering
    void renderQuad();
    void applyEffects();

    // Shader programs
    std::unique_ptr<QOpenGLShaderProgram> m_basicShader;
    std::unique_ptr<QOpenGLShaderProgram> m_effectShader;

    // OpenGL objects
    std::unique_ptr<QOpenGLTexture> m_texture;
    std::unique_ptr<QOpenGLBuffer> m_vertexBuffer;
    std::unique_ptr<QOpenGLBuffer> m_indexBuffer;
    std::unique_ptr<QOpenGLVertexArrayObject> m_vao;

    // Effect parameters
    bool m_effectsEnabled;
    float m_brightness;
    float m_contrast;
    float m_saturation;

    // Current frame
    QImage m_currentImage;
    bool m_hasFrame;

    // Vertex data
    struct Vertex {
        float position[3];
        float texCoord[2];
    };
};

#endif // GL_VIDEO_RENDERER_H

