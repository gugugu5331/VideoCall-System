#version 330 core

in vec2 TexCoord;
out vec4 FragColor;

uniform sampler2D videoTexture;
uniform float time;
uniform vec2 resolution;

// 基础参数
uniform float brightness;
uniform float contrast;
uniform float saturation;
uniform float hue;
uniform float gamma;
uniform vec3 colorBalance;

// 滤镜参数
uniform int filterType;
uniform float filterIntensity;

// HSV转换函数
vec3 rgb2hsv(vec3 c) {
    vec4 K = vec4(0.0, -1.0 / 3.0, 2.0 / 3.0, -1.0);
    vec4 p = mix(vec4(c.bg, K.wz), vec4(c.gb, K.xy), step(c.b, c.g));
    vec4 q = mix(vec4(p.xyw, c.r), vec4(c.r, p.yzx), step(p.x, c.r));
    
    float d = q.x - min(q.w, q.y);
    float e = 1.0e-10;
    return vec3(abs(q.z + (q.w - q.y) / (6.0 * d + e)), d / (q.x + e), q.x);
}

vec3 hsv2rgb(vec3 c) {
    vec4 K = vec4(1.0, 2.0 / 3.0, 1.0 / 3.0, 3.0);
    vec3 p = abs(fract(c.xxx + K.xyz) * 6.0 - K.www);
    return c.z * mix(K.xxx, clamp(p - K.xxx, 0.0, 1.0), c.y);
}

// 滤镜效果函数
vec3 applyBlur(vec2 uv) {
    vec3 color = vec3(0.0);
    float kernel[9] = float[](
        1.0/16.0, 2.0/16.0, 1.0/16.0,
        2.0/16.0, 4.0/16.0, 2.0/16.0,
        1.0/16.0, 2.0/16.0, 1.0/16.0
    );
    
    vec2 offset = 1.0 / resolution;
    for (int i = -1; i <= 1; i++) {
        for (int j = -1; j <= 1; j++) {
            vec2 sampleUV = uv + vec2(float(i), float(j)) * offset;
            color += texture(videoTexture, sampleUV).rgb * kernel[(i+1)*3 + (j+1)];
        }
    }
    return color;
}

vec3 applySharpen(vec2 uv) {
    vec3 color = vec3(0.0);
    float kernel[9] = float[](
        0.0, -1.0, 0.0,
        -1.0, 5.0, -1.0,
        0.0, -1.0, 0.0
    );
    
    vec2 offset = 1.0 / resolution;
    for (int i = -1; i <= 1; i++) {
        for (int j = -1; j <= 1; j++) {
            vec2 sampleUV = uv + vec2(float(i), float(j)) * offset;
            color += texture(videoTexture, sampleUV).rgb * kernel[(i+1)*3 + (j+1)];
        }
    }
    return color;
}

vec3 applyEdgeDetection(vec2 uv) {
    vec3 color = vec3(0.0);
    float kernel[9] = float[](
        -1.0, -1.0, -1.0,
        -1.0, 8.0, -1.0,
        -1.0, -1.0, -1.0
    );
    
    vec2 offset = 1.0 / resolution;
    for (int i = -1; i <= 1; i++) {
        for (int j = -1; j <= 1; j++) {
            vec2 sampleUV = uv + vec2(float(i), float(j)) * offset;
            color += texture(videoTexture, sampleUV).rgb * kernel[(i+1)*3 + (j+1)];
        }
    }
    return color;
}

vec3 applySepia(vec3 color) {
    mat3 sepiaMatrix = mat3(
        0.393, 0.769, 0.189,
        0.349, 0.686, 0.168,
        0.272, 0.534, 0.131
    );
    return sepiaMatrix * color;
}

vec3 applyVintage(vec3 color) {
    // 降低饱和度
    vec3 hsv = rgb2hsv(color);
    hsv.y *= 0.7;
    color = hsv2rgb(hsv);
    
    // 添加暖色调
    color.r *= 1.1;
    color.g *= 1.05;
    color.b *= 0.9;
    
    // 添加渐晕效果
    vec2 center = vec2(0.5, 0.5);
    float dist = distance(TexCoord, center);
    float vignette = 1.0 - smoothstep(0.3, 0.8, dist);
    color *= vignette;
    
    return color;
}

vec3 applyCartoon(vec3 color) {
    // 量化颜色
    color = floor(color * 8.0) / 8.0;
    
    // 增强对比度
    color = pow(color, vec3(0.8));
    
    return color;
}

void main() {
    vec2 uv = TexCoord;
    vec3 color = texture(videoTexture, uv).rgb;
    
    // 应用滤镜
    if (filterType == 1) { // BLUR
        color = mix(color, applyBlur(uv), filterIntensity);
    } else if (filterType == 2) { // SHARPEN
        color = mix(color, applySharpen(uv), filterIntensity);
    } else if (filterType == 3) { // EDGE_DETECTION
        color = mix(color, applyEdgeDetection(uv), filterIntensity);
    } else if (filterType == 5) { // SEPIA
        color = mix(color, applySepia(color), filterIntensity);
    } else if (filterType == 6) { // VINTAGE
        color = mix(color, applyVintage(color), filterIntensity);
    } else if (filterType == 8) { // CARTOON
        color = mix(color, applyCartoon(color), filterIntensity);
    }
    
    // 应用基础调整
    color *= colorBalance;
    color = (color - 0.5) * contrast + 0.5 + brightness;
    
    // 饱和度调整
    vec3 hsv = rgb2hsv(color);
    hsv.y *= saturation;
    hsv.x += hue;
    color = hsv2rgb(hsv);
    
    // Gamma校正
    color = pow(color, vec3(1.0 / gamma));
    
    // 确保颜色在有效范围内
    color = clamp(color, 0.0, 1.0);
    
    FragColor = vec4(color, 1.0);
}
