#include "video_effects/gl_video_renderer.h"
#include <QDebug>
#include <QOpenGLContext>

// Vertex shader source
static const char *vertexShaderSource = R"(
#version 330 core
layout (location = 0) in vec3 aPos;
layout (location = 1) in vec2 aTexCoord;

out vec2 TexCoord;

void main()
{
    gl_Position = vec4(aPos, 1.0);
    TexCoord = aTexCoord;
}
)";

// Basic fragment shader (no effects)
static const char *basicFragmentShaderSource = R"(
#version 330 core
out vec4 FragColor;

in vec2 TexCoord;

uniform sampler2D texture1;

void main()
{
    FragColor = texture(texture1, TexCoord);
}
)";

// Effect fragment shader (with brightness, contrast, saturation)
static const char *effectFragmentShaderSource = R"(
#version 330 core
out vec4 FragColor;

in vec2 TexCoord;

uniform sampler2D texture1;
uniform float brightness;
uniform float contrast;
uniform float saturation;

vec3 adjustBrightness(vec3 color, float value) {
    return color + value;
}

vec3 adjustContrast(vec3 color, float value) {
    return (color - 0.5) * value + 0.5;
}

vec3 adjustSaturation(vec3 color, float value) {
    float gray = dot(color, vec3(0.299, 0.587, 0.114));
    return mix(vec3(gray), color, value);
}

void main()
{
    vec4 texColor = texture(texture1, TexCoord);
    vec3 color = texColor.rgb;
    
    // Apply effects
    color = adjustBrightness(color, brightness);
    color = adjustContrast(color, contrast);
    color = adjustSaturation(color, saturation);
    
    // Clamp to [0, 1]
    color = clamp(color, 0.0, 1.0);
    
    FragColor = vec4(color, texColor.a);
}
)";

GLVideoRenderer::GLVideoRenderer(QWidget *parent)
    : QOpenGLWidget(parent)
    , m_effectsEnabled(false)
    , m_brightness(0.0f)
    , m_contrast(1.0f)
    , m_saturation(1.0f)
    , m_hasFrame(false)
{
    // Enable OpenGL context
    QSurfaceFormat format;
    format.setVersion(3, 3);
    format.setProfile(QSurfaceFormat::CoreProfile);
    format.setDepthBufferSize(24);
    format.setStencilBufferSize(8);
    setFormat(format);
}

GLVideoRenderer::~GLVideoRenderer()
{
    makeCurrent();
    deleteTexture();
    m_vao.reset();
    m_vertexBuffer.reset();
    m_indexBuffer.reset();
    m_basicShader.reset();
    m_effectShader.reset();
    doneCurrent();
}

void GLVideoRenderer::setEffectsEnabled(bool enabled)
{
    if (m_effectsEnabled != enabled) {
        m_effectsEnabled = enabled;
        emit effectsEnabledChanged();
        update();
    }
}

void GLVideoRenderer::setBrightness(float value)
{
    value = qBound(-1.0f, value, 1.0f);
    if (qAbs(m_brightness - value) > 0.001f) {
        m_brightness = value;
        emit brightnessChanged();
        update();
    }
}

void GLVideoRenderer::setContrast(float value)
{
    value = qBound(0.0f, value, 2.0f);
    if (qAbs(m_contrast - value) > 0.001f) {
        m_contrast = value;
        emit contrastChanged();
        update();
    }
}

void GLVideoRenderer::setSaturation(float value)
{
    value = qBound(0.0f, value, 2.0f);
    if (qAbs(m_saturation - value) > 0.001f) {
        m_saturation = value;
        emit saturationChanged();
        update();
    }
}

void GLVideoRenderer::setVideoFrame(const QVideoFrame &frame)
{
    if (!frame.isValid()) {
        return;
    }

    QVideoFrame f = frame;
    if (!f.map(QVideoFrame::ReadOnly)) {
        qWarning() << "Failed to map video frame";
        return;
    }

    QImage image = f.toImage();
    f.unmap();

    setImage(image);
}

void GLVideoRenderer::setImage(const QImage &image)
{
    if (image.isNull()) {
        return;
    }

    m_currentImage = image.convertToFormat(QImage::Format_RGBA8888);
    m_hasFrame = true;

    makeCurrent();
    updateTexture(m_currentImage);
    doneCurrent();

    update();
}

void GLVideoRenderer::clear()
{
    m_hasFrame = false;
    m_currentImage = QImage();
    
    makeCurrent();
    deleteTexture();
    doneCurrent();
    
    update();
}

void GLVideoRenderer::initializeGL()
{
    initializeOpenGLFunctions();

    // Set clear color
    glClearColor(0.0f, 0.0f, 0.0f, 1.0f);

    // Initialize shaders
    if (!initShaders()) {
        qCritical() << "Failed to initialize shaders";
        emit renderError("Failed to initialize shaders");
        return;
    }

    // Create vertex buffer
    m_vertexBuffer = std::make_unique<QOpenGLBuffer>(QOpenGLBuffer::VertexBuffer);
    m_vertexBuffer->create();
    m_vertexBuffer->bind();

    // Vertex data (position + texcoord)
    Vertex vertices[] = {
        // positions          // texture coords
        {{-1.0f,  1.0f, 0.0f}, {0.0f, 0.0f}},  // top left
        {{-1.0f, -1.0f, 0.0f}, {0.0f, 1.0f}},  // bottom left
        {{ 1.0f, -1.0f, 0.0f}, {1.0f, 1.0f}},  // bottom right
        {{ 1.0f,  1.0f, 0.0f}, {1.0f, 0.0f}}   // top right
    };

    m_vertexBuffer->allocate(vertices, sizeof(vertices));
    m_vertexBuffer->release();

    // Create index buffer
    m_indexBuffer = std::make_unique<QOpenGLBuffer>(QOpenGLBuffer::IndexBuffer);
    m_indexBuffer->create();
    m_indexBuffer->bind();

    unsigned int indices[] = {
        0, 1, 2,  // first triangle
        0, 2, 3   // second triangle
    };

    m_indexBuffer->allocate(indices, sizeof(indices));
    m_indexBuffer->release();

    // Create VAO
    m_vao = std::make_unique<QOpenGLVertexArrayObject>();
    m_vao->create();
    m_vao->bind();

    m_vertexBuffer->bind();
    m_indexBuffer->bind();

    // Position attribute
    glEnableVertexAttribArray(0);
    glVertexAttribPointer(0, 3, GL_FLOAT, GL_FALSE, sizeof(Vertex), (void*)0);

    // Texture coordinate attribute
    glEnableVertexAttribArray(1);
    glVertexAttribPointer(1, 2, GL_FLOAT, GL_FALSE, sizeof(Vertex), (void*)offsetof(Vertex, texCoord));

    m_vao->release();

    qDebug() << "OpenGL initialized successfully";
}

void GLVideoRenderer::resizeGL(int w, int h)
{
    glViewport(0, 0, w, h);
}

void GLVideoRenderer::paintGL()
{
    glClear(GL_COLOR_BUFFER_BIT);

    if (!m_hasFrame || !m_texture) {
        return;
    }

    // Choose shader based on effects enabled
    QOpenGLShaderProgram *shader = m_effectsEnabled ? m_effectShader.get() : m_basicShader.get();
    
    if (!shader || !shader->bind()) {
        return;
    }

    // Set uniforms
    if (m_effectsEnabled) {
        shader->setUniformValue("brightness", m_brightness);
        shader->setUniformValue("contrast", m_contrast);
        shader->setUniformValue("saturation", m_saturation);
    }

    shader->setUniformValue("texture1", 0);

    // Bind texture
    glActiveTexture(GL_TEXTURE0);
    m_texture->bind();

    // Draw quad
    m_vao->bind();
    glDrawElements(GL_TRIANGLES, 6, GL_UNSIGNED_INT, 0);
    m_vao->release();

    m_texture->release();
    shader->release();
}

bool GLVideoRenderer::initShaders()
{
    // Create basic shader
    m_basicShader = std::make_unique<QOpenGLShaderProgram>();
    if (!m_basicShader->addShaderFromSourceCode(QOpenGLShader::Vertex, vertexShaderSource)) {
        qCritical() << "Failed to compile basic vertex shader:" << m_basicShader->log();
        return false;
    }
    if (!m_basicShader->addShaderFromSourceCode(QOpenGLShader::Fragment, basicFragmentShaderSource)) {
        qCritical() << "Failed to compile basic fragment shader:" << m_basicShader->log();
        return false;
    }
    if (!m_basicShader->link()) {
        qCritical() << "Failed to link basic shader:" << m_basicShader->log();
        return false;
    }

    // Create effect shader
    m_effectShader = std::make_unique<QOpenGLShaderProgram>();
    if (!m_effectShader->addShaderFromSourceCode(QOpenGLShader::Vertex, vertexShaderSource)) {
        qCritical() << "Failed to compile effect vertex shader:" << m_effectShader->log();
        return false;
    }
    if (!m_effectShader->addShaderFromSourceCode(QOpenGLShader::Fragment, effectFragmentShaderSource)) {
        qCritical() << "Failed to compile effect fragment shader:" << m_effectShader->log();
        return false;
    }
    if (!m_effectShader->link()) {
        qCritical() << "Failed to link effect shader:" << m_effectShader->log();
        return false;
    }

    return true;
}

void GLVideoRenderer::updateTexture(const QImage &image)
{
    if (image.isNull()) {
        return;
    }

    if (!m_texture) {
        m_texture = std::make_unique<QOpenGLTexture>(QOpenGLTexture::Target2D);
        m_texture->create();
    }

    m_texture->destroy();
    m_texture->create();
    m_texture->setData(image.mirrored());
    m_texture->setMinificationFilter(QOpenGLTexture::Linear);
    m_texture->setMagnificationFilter(QOpenGLTexture::Linear);
    m_texture->setWrapMode(QOpenGLTexture::ClampToEdge);
}

void GLVideoRenderer::deleteTexture()
{
    if (m_texture) {
        m_texture->destroy();
        m_texture.reset();
    }
}

