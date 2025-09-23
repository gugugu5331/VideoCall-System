#!/usr/bin/env python3
# -*- coding: utf-8 -*-
"""
FFmpeg服务项目集成脚本
将FFmpeg服务集成到现有的VideoCall项目中
"""

import os
import sys
import json
import shutil
import subprocess
from pathlib import Path

def run_command(command, capture_output=True):
    """运行命令并返回结果"""
    try:
        result = subprocess.run(
            command,
            shell=True,
            capture_output=capture_output,
            text=True,
            encoding='utf-8',
            errors='ignore'
        )
        return result
    except Exception as e:
        print(f"运行命令错误 '{command}': {e}")
        return None

def check_file_exists(file_path):
    """检查文件是否存在"""
    return os.path.exists(file_path)

def integrate_with_python_ai_service():
    """集成到Python AI服务"""
    print("=" * 60)
    print("集成到Python AI服务")
    print("=" * 60)
    
    # 检查AI服务目录
    ai_service_dir = Path("../../ai-service")
    if not ai_service_dir.exists():
        print("❌ AI服务目录不存在")
        return False
    
    # 创建FFmpeg服务集成文件
    integration_file = ai_service_dir / "app" / "services" / "ffmpeg_integration.py"
    integration_file.parent.mkdir(parents=True, exist_ok=True)
    
    integration_code = '''#!/usr/bin/env python3
# -*- coding: utf-8 -*-
"""
FFmpeg服务集成模块
提供Python接口调用C++ FFmpeg服务
"""

import os
import json
import subprocess
import tempfile
from pathlib import Path
from typing import Dict, List, Optional, Union
import base64

class FFmpegServiceIntegration:
    """FFmpeg服务集成类"""
    
    def __init__(self, service_path: str = None):
        """
        初始化FFmpeg服务集成
        
        Args:
            service_path: FFmpeg服务可执行文件路径
        """
        if service_path is None:
            # 默认路径
            current_dir = Path(__file__).parent.parent.parent.parent
            service_path = current_dir / "ffmpeg-service" / "build" / "bin" / "ffmpeg_service_example"
            if os.name == 'nt':  # Windows
                service_path = service_path.with_suffix('.exe')
        
        self.service_path = Path(service_path)
        if not self.service_path.exists():
            raise FileNotFoundError(f"FFmpeg服务未找到: {self.service_path}")
    
    def detect_video_forgery(self, video_data: bytes, config: Dict = None) -> Dict:
        """
        检测视频伪造
        
        Args:
            video_data: 视频数据
            config: 检测配置
            
        Returns:
            检测结果
        """
        try:
            # 创建临时文件
            with tempfile.NamedTemporaryFile(suffix='.mp4', delete=False) as f:
                f.write(video_data)
                temp_video = f.name
            
            # 准备配置
            if config is None:
                config = {
                    "detection_type": "video_deepfake",
                    "confidence_threshold": 0.8,
                    "processing_mode": "fast"
                }
            
            config_file = temp_video + ".config.json"
            with open(config_file, 'w') as f:
                json.dump(config, f)
            
            try:
                # 调用FFmpeg服务
                cmd = [str(self.service_path), "detect_video", temp_video, config_file]
                result = subprocess.run(cmd, capture_output=True, text=True, timeout=60)
                
                if result.returncode != 0:
                    return {
                        "success": False,
                        "error": f"FFmpeg服务执行失败: {result.stderr}"
                    }
                
                # 解析结果
                try:
                    output = json.loads(result.stdout)
                    return {
                        "success": True,
                        "result": output
                    }
                except json.JSONDecodeError:
                    return {
                        "success": False,
                        "error": f"无法解析FFmpeg服务输出: {result.stdout}"
                    }
                    
            finally:
                # 清理临时文件
                if os.path.exists(temp_video):
                    os.unlink(temp_video)
                if os.path.exists(config_file):
                    os.unlink(config_file)
                    
        except Exception as e:
            return {
                "success": False,
                "error": f"视频伪造检测异常: {str(e)}"
            }
    
    def detect_audio_forgery(self, audio_data: bytes, config: Dict = None) -> Dict:
        """
        检测音频伪造
        
        Args:
            audio_data: 音频数据
            config: 检测配置
            
        Returns:
            检测结果
        """
        try:
            # 创建临时文件
            with tempfile.NamedTemporaryFile(suffix='.wav', delete=False) as f:
                f.write(audio_data)
                temp_audio = f.name
            
            # 准备配置
            if config is None:
                config = {
                    "detection_type": "voice_spoofing",
                    "confidence_threshold": 0.8,
                    "processing_mode": "fast"
                }
            
            config_file = temp_audio + ".config.json"
            with open(config_file, 'w') as f:
                json.dump(config, f)
            
            try:
                # 调用FFmpeg服务
                cmd = [str(self.service_path), "detect_audio", temp_audio, config_file]
                result = subprocess.run(cmd, capture_output=True, text=True, timeout=60)
                
                if result.returncode != 0:
                    return {
                        "success": False,
                        "error": f"FFmpeg服务执行失败: {result.stderr}"
                    }
                
                # 解析结果
                try:
                    output = json.loads(result.stdout)
                    return {
                        "success": True,
                        "result": output
                    }
                except json.JSONDecodeError:
                    return {
                        "success": False,
                        "error": f"无法解析FFmpeg服务输出: {result.stdout}"
                    }
                    
            finally:
                # 清理临时文件
                if os.path.exists(temp_audio):
                    os.unlink(temp_audio)
                if os.path.exists(config_file):
                    os.unlink(config_file)
                    
        except Exception as e:
            return {
                "success": False,
                "error": f"音频伪造检测异常: {str(e)}"
            }
    
    def compress_media(self, media_data: bytes, media_type: str, config: Dict = None) -> Dict:
        """
        压缩媒体文件
        
        Args:
            media_data: 媒体数据
            media_type: 媒体类型 ('video' 或 'audio')
            config: 压缩配置
            
        Returns:
            压缩结果
        """
        try:
            # 确定文件扩展名
            if media_type == 'video':
                suffix = '.mp4'
            elif media_type == 'audio':
                suffix = '.wav'
            else:
                return {
                    "success": False,
                    "error": f"不支持的媒体类型: {media_type}"
                }
            
            # 创建临时文件
            with tempfile.NamedTemporaryFile(suffix=suffix, delete=False) as f:
                f.write(media_data)
                temp_media = f.name
            
            # 准备配置
            if config is None:
                config = {
                    "compression_level": "medium",
                    "quality": 0.8,
                    "format": "mp4" if media_type == 'video' else "aac"
                }
            
            config_file = temp_media + ".config.json"
            with open(config_file, 'w') as f:
                json.dump(config, f)
            
            try:
                # 调用FFmpeg服务
                cmd = [str(self.service_path), "compress", temp_media, config_file]
                result = subprocess.run(cmd, capture_output=True, text=True, timeout=120)
                
                if result.returncode != 0:
                    return {
                        "success": False,
                        "error": f"FFmpeg服务执行失败: {result.stderr}"
                    }
                
                # 解析结果
                try:
                    output = json.loads(result.stdout)
                    return {
                        "success": True,
                        "result": output
                    }
                except json.JSONDecodeError:
                    return {
                        "success": False,
                        "error": f"无法解析FFmpeg服务输出: {result.stdout}"
                    }
                    
            finally:
                # 清理临时文件
                if os.path.exists(temp_media):
                    os.unlink(temp_media)
                if os.path.exists(config_file):
                    os.unlink(config_file)
                    
        except Exception as e:
            return {
                "success": False,
                "error": f"媒体压缩异常: {str(e)}"
            }

# 全局实例
ffmpeg_service = None

def get_ffmpeg_service() -> FFmpegServiceIntegration:
    """获取FFmpeg服务实例"""
    global ffmpeg_service
    if ffmpeg_service is None:
        ffmpeg_service = FFmpegServiceIntegration()
    return ffmpeg_service
'''
    
    with open(integration_file, 'w', encoding='utf-8') as f:
        f.write(integration_code)
    
    print(f"✅ FFmpeg服务集成文件已创建: {integration_file}")
    return True

def integrate_with_go_backend():
    """集成到Go后端"""
    print("\n" + "=" * 60)
    print("集成到Go后端")
    print("=" * 60)
    
    # 检查Go后端目录
    backend_dir = Path("../../backend")
    if not backend_dir.exists():
        print("❌ Go后端目录不存在")
        return False
    
    # 创建FFmpeg服务集成文件
    integration_file = backend_dir / "handlers" / "ffmpeg_handler.go"
    integration_file.parent.mkdir(parents=True, exist_ok=True)
    
    integration_code = '''package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"time"
)

// FFmpegServiceConfig FFmpeg服务配置
type FFmpegServiceConfig struct {
	ServicePath string `json:"service_path"`
	Timeout     int    `json:"timeout"`
}

// FFmpegDetectionRequest 检测请求
type FFmpegDetectionRequest struct {
	MediaData    []byte                 `json:"media_data"`
	MediaType    string                 `json:"media_type"`
	DetectionType string                `json:"detection_type"`
	Config       map[string]interface{} `json:"config"`
}

// FFmpegDetectionResponse 检测响应
type FFmpegDetectionResponse struct {
	Success bool                   `json:"success"`
	Result  map[string]interface{} `json:"result,omitempty"`
	Error   string                 `json:"error,omitempty"`
}

// FFmpegHandler FFmpeg服务处理器
type FFmpegHandler struct {
	config FFmpegServiceConfig
}

// NewFFmpegHandler 创建FFmpeg处理器
func NewFFmpegHandler(config FFmpegServiceConfig) *FFmpegHandler {
	return &FFmpegHandler{
		config: config,
	}
}

// DetectForgery 检测伪造
func (h *FFmpegHandler) DetectForgery(w http.ResponseWriter, r *http.Request) {
	var req FFmpegDetectionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// 创建临时文件
	tempFile, err := os.CreateTemp("", "ffmpeg_*")
	if err != nil {
		http.Error(w, "Failed to create temp file", http.StatusInternalServerError)
		return
	}
	defer os.Remove(tempFile.Name())

	// 写入媒体数据
	if _, err := tempFile.Write(req.MediaData); err != nil {
		http.Error(w, "Failed to write temp file", http.StatusInternalServerError)
		return
	}
	tempFile.Close()

	// 准备配置
	config := req.Config
	if config == nil {
		config = map[string]interface{}{
			"detection_type": req.DetectionType,
			"confidence_threshold": 0.8,
			"processing_mode": "fast",
		}
	}

	configFile := tempFile.Name() + ".config.json"
	configData, err := json.Marshal(config)
	if err != nil {
		http.Error(w, "Failed to marshal config", http.StatusInternalServerError)
		return
	}

	if err := os.WriteFile(configFile, configData, 0644); err != nil {
		http.Error(w, "Failed to write config file", http.StatusInternalServerError)
		return
	}
	defer os.Remove(configFile)

	// 调用FFmpeg服务
	cmd := exec.Command(h.config.ServicePath, "detect", tempFile.Name(), configFile)
	cmd.Timeout = time.Duration(h.config.Timeout) * time.Second

	output, err := cmd.CombinedOutput()
	if err != nil {
		response := FFmpegDetectionResponse{
			Success: false,
			Error:   fmt.Sprintf("FFmpeg service execution failed: %v", err),
		}
		json.NewEncoder(w).Encode(response)
		return
	}

	// 解析结果
	var result map[string]interface{}
	if err := json.Unmarshal(output, &result); err != nil {
		response := FFmpegDetectionResponse{
			Success: false,
			Error:   fmt.Sprintf("Failed to parse FFmpeg service output: %v", err),
		}
		json.NewEncoder(w).Encode(response)
		return
	}

	response := FFmpegDetectionResponse{
		Success: true,
		Result:  result,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// CompressMedia 压缩媒体
func (h *FFmpegHandler) CompressMedia(w http.ResponseWriter, r *http.Request) {
	var req FFmpegDetectionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// 创建临时文件
	tempFile, err := os.CreateTemp("", "ffmpeg_*")
	if err != nil {
		http.Error(w, "Failed to create temp file", http.StatusInternalServerError)
		return
	}
	defer os.Remove(tempFile.Name())

	// 写入媒体数据
	if _, err := tempFile.Write(req.MediaData); err != nil {
		http.Error(w, "Failed to write temp file", http.StatusInternalServerError)
		return
	}
	tempFile.Close()

	// 准备配置
	config := req.Config
	if config == nil {
		config = map[string]interface{}{
			"compression_level": "medium",
			"quality": 0.8,
			"format": "mp4",
		}
	}

	configFile := tempFile.Name() + ".config.json"
	configData, err := json.Marshal(config)
	if err != nil {
		http.Error(w, "Failed to marshal config", http.StatusInternalServerError)
		return
	}

	if err := os.WriteFile(configFile, configData, 0644); err != nil {
		http.Error(w, "Failed to write config file", http.StatusInternalServerError)
		return
	}
	defer os.Remove(configFile)

	// 调用FFmpeg服务
	cmd := exec.Command(h.config.ServicePath, "compress", tempFile.Name(), configFile)
	cmd.Timeout = time.Duration(h.config.Timeout) * time.Second

	output, err := cmd.CombinedOutput()
	if err != nil {
		response := FFmpegDetectionResponse{
			Success: false,
			Error:   fmt.Sprintf("FFmpeg service execution failed: %v", err),
		}
		json.NewEncoder(w).Encode(response)
		return
	}

	// 解析结果
	var result map[string]interface{}
	if err := json.Unmarshal(output, &result); err != nil {
		response := FFmpegDetectionResponse{
			Success: false,
			Error:   fmt.Sprintf("Failed to parse FFmpeg service output: %v", err),
		}
		json.NewEncoder(w).Encode(response)
		return
	}

	response := FFmpegDetectionResponse{
		Success: true,
		Result:  result,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
'''
    
    with open(integration_file, 'w', encoding='utf-8') as f:
        f.write(integration_code)
    
    print(f"✅ Go后端集成文件已创建: {integration_file}")
    return True

def integrate_with_webrtc():
    """集成到WebRTC前端"""
    print("\n" + "=" * 60)
    print("集成到WebRTC前端")
    print("=" * 60)
    
    # 检查WebRTC前端目录
    webrtc_dir = Path("../../web_interface")
    if not webrtc_dir.exists():
        print("❌ WebRTC前端目录不存在")
        return False
    
    # 创建FFmpeg服务集成文件
    integration_file = webrtc_dir / "js" / "ffmpeg-integration.js"
    integration_file.parent.mkdir(parents=True, exist_ok=True)
    
    integration_code = '''/**
 * FFmpeg服务前端集成
 * 提供WebRTC音视频流的实时检测和压缩功能
 */

class FFmpegServiceIntegration {
    constructor(serviceUrl = '/api/ffmpeg') {
        this.serviceUrl = serviceUrl;
        this.isInitialized = false;
    }

    /**
     * 初始化FFmpeg服务
     */
    async initialize() {
        try {
            // 检查服务可用性
            const response = await fetch(`${this.serviceUrl}/health`);
            if (!response.ok) {
                throw new Error('FFmpeg服务不可用');
            }
            
            this.isInitialized = true;
            console.log('FFmpeg服务初始化成功');
            return true;
        } catch (error) {
            console.error('FFmpeg服务初始化失败:', error);
            return false;
        }
    }

    /**
     * 检测视频流伪造
     * @param {MediaStream} videoStream 视频流
     * @param {Object} config 检测配置
     * @returns {Promise<Object>} 检测结果
     */
    async detectVideoForgery(videoStream, config = {}) {
        if (!this.isInitialized) {
            throw new Error('FFmpeg服务未初始化');
        }

        try {
            // 从视频流中提取帧
            const videoTrack = videoStream.getVideoTracks()[0];
            const imageCapture = new ImageCapture(videoTrack);
            const bitmap = await imageCapture.grabFrame();

            // 转换为Blob
            const canvas = document.createElement('canvas');
            canvas.width = bitmap.width;
            canvas.height = bitmap.height;
            const ctx = canvas.getContext('2d');
            ctx.drawImage(bitmap, 0, 0);
            
            const blob = await new Promise(resolve => {
                canvas.toBlob(resolve, 'image/jpeg', 0.9);
            });

            // 发送到后端进行检测
            const formData = new FormData();
            formData.append('video_frame', blob);
            formData.append('config', JSON.stringify(config));

            const response = await fetch(`${this.serviceUrl}/detect-video`, {
                method: 'POST',
                body: formData
            });

            if (!response.ok) {
                throw new Error(`检测请求失败: ${response.status}`);
            }

            const result = await response.json();
            return result;
        } catch (error) {
            console.error('视频伪造检测失败:', error);
            throw error;
        }
    }

    /**
     * 检测音频流伪造
     * @param {MediaStream} audioStream 音频流
     * @param {Object} config 检测配置
     * @returns {Promise<Object>} 检测结果
     */
    async detectAudioForgery(audioStream, config = {}) {
        if (!this.isInitialized) {
            throw new Error('FFmpeg服务未初始化');
        }

        try {
            // 录制音频片段
            const mediaRecorder = new MediaRecorder(audioStream, {
                mimeType: 'audio/webm;codecs=opus'
            });

            const audioChunks = [];
            
            return new Promise((resolve, reject) => {
                mediaRecorder.ondataavailable = (event) => {
                    audioChunks.push(event.data);
                };

                mediaRecorder.onstop = async () => {
                    try {
                        const audioBlob = new Blob(audioChunks, { type: 'audio/webm' });
                        
                        // 发送到后端进行检测
                        const formData = new FormData();
                        formData.append('audio_data', audioBlob);
                        formData.append('config', JSON.stringify(config));

                        const response = await fetch(`${this.serviceUrl}/detect-audio`, {
                            method: 'POST',
                            body: formData
                        });

                        if (!response.ok) {
                            throw new Error(`检测请求失败: ${response.status}`);
                        }

                        const result = await response.json();
                        resolve(result);
                    } catch (error) {
                        reject(error);
                    }
                };

                // 录制3秒音频
                mediaRecorder.start();
                setTimeout(() => {
                    mediaRecorder.stop();
                }, 3000);
            });
        } catch (error) {
            console.error('音频伪造检测失败:', error);
            throw error;
        }
    }

    /**
     * 压缩视频流
     * @param {MediaStream} videoStream 视频流
     * @param {Object} config 压缩配置
     * @returns {Promise<MediaStream>} 压缩后的视频流
     */
    async compressVideoStream(videoStream, config = {}) {
        if (!this.isInitialized) {
            throw new Error('FFmpeg服务未初始化');
        }

        try {
            // 创建MediaRecorder进行压缩
            const compressionConfig = {
                video: {
                    codec: 'h264',
                    bitrate: config.bitrate || 1000000, // 1Mbps
                    framerate: config.framerate || 30,
                    width: config.width || 1280,
                    height: config.height || 720
                },
                audio: {
                    codec: 'aac',
                    bitrate: config.audioBitrate || 128000 // 128kbps
                }
            };

            const mediaRecorder = new MediaRecorder(videoStream, {
                mimeType: 'video/webm;codecs=h264'
            });

            // 创建新的MediaStream
            const canvas = document.createElement('canvas');
            const ctx = canvas.getContext('2d');
            const video = document.createElement('video');
            
            video.srcObject = videoStream;
            await video.play();

            canvas.width = compressionConfig.video.width;
            canvas.height = compressionConfig.video.height;

            // 创建压缩后的流
            const compressedStream = canvas.captureStream(compressionConfig.video.framerate);
            
            // 添加音频轨道
            const audioTracks = videoStream.getAudioTracks();
            if (audioTracks.length > 0) {
                compressedStream.addTrack(audioTracks[0]);
            }

            return compressedStream;
        } catch (error) {
            console.error('视频流压缩失败:', error);
            throw error;
        }
    }

    /**
     * 实时检测监控
     * @param {MediaStream} mediaStream 媒体流
     * @param {Function} callback 检测结果回调
     * @param {Object} config 配置
     */
    startRealTimeDetection(mediaStream, callback, config = {}) {
        if (!this.isInitialized) {
            throw new Error('FFmpeg服务未初始化');
        }

        const interval = config.interval || 5000; // 5秒检测一次
        const detectionConfig = config.detection || {};

        const detectionInterval = setInterval(async () => {
            try {
                const videoResult = await this.detectVideoForgery(mediaStream, detectionConfig);
                const audioResult = await this.detectAudioForgery(mediaStream, detectionConfig);

                callback({
                    video: videoResult,
                    audio: audioResult,
                    timestamp: Date.now()
                });
            } catch (error) {
                console.error('实时检测失败:', error);
                callback({
                    error: error.message,
                    timestamp: Date.now()
                });
            }
        }, interval);

        // 返回停止函数
        return () => {
            clearInterval(detectionInterval);
        };
    }
}

// 全局实例
window.ffmpegService = new FFmpegServiceIntegration();

// 导出模块
if (typeof module !== 'undefined' && module.exports) {
    module.exports = FFmpegServiceIntegration;
}
'''
    
    with open(integration_file, 'w', encoding='utf-8') as f:
        f.write(integration_code)
    
    print(f"✅ WebRTC前端集成文件已创建: {integration_file}")
    return True

def create_integration_config():
    """创建集成配置文件"""
    print("\n" + "=" * 60)
    print("创建集成配置文件")
    print("=" * 60)
    
    config = {
        "ffmpeg_service": {
            "enabled": True,
            "service_path": "./core/ffmpeg-service/build/bin/ffmpeg_service_example",
            "timeout": 60,
            "max_concurrent_requests": 10
        },
        "integration": {
            "python_ai_service": {
                "enabled": True,
                "endpoint": "/api/ffmpeg",
                "methods": ["detect_video", "detect_audio", "compress_media"]
            },
            "go_backend": {
                "enabled": True,
                "endpoint": "/api/ffmpeg",
                "methods": ["detect_forgery", "compress_media"]
            },
            "webrtc_frontend": {
                "enabled": True,
                "real_time_detection": True,
                "compression_enabled": True
            }
        },
        "detection": {
            "video_deepfake": {
                "enabled": True,
                "confidence_threshold": 0.8,
                "processing_mode": "fast"
            },
            "voice_spoofing": {
                "enabled": True,
                "confidence_threshold": 0.8,
                "processing_mode": "fast"
            },
            "face_swap": {
                "enabled": True,
                "confidence_threshold": 0.8,
                "processing_mode": "fast"
            }
        },
        "compression": {
            "video": {
                "codec": "h264",
                "bitrate": 1000000,
                "quality": 0.8
            },
            "audio": {
                "codec": "aac",
                "bitrate": 128000,
                "quality": 0.8
            }
        }
    }
    
    config_file = Path("integration_config.json")
    with open(config_file, 'w', encoding='utf-8') as f:
        json.dump(config, f, indent=2, ensure_ascii=False)
    
    print(f"✅ 集成配置文件已创建: {config_file}")
    return True

def main():
    """主函数"""
    print("FFmpeg服务项目集成")
    print("=" * 60)
    
    integrations = [
        ("Python AI服务集成", integrate_with_python_ai_service),
        ("Go后端集成", integrate_with_go_backend),
        ("WebRTC前端集成", integrate_with_webrtc),
        ("集成配置", create_integration_config)
    ]
    
    results = []
    for integration_name, integration_func in integrations:
        try:
            result = integration_func()
            results.append((integration_name, result))
        except Exception as e:
            print(f"❌ {integration_name}异常: {e}")
            results.append((integration_name, False))
    
    # 总结
    print("\n" + "=" * 60)
    print("集成总结")
    print("=" * 60)
    
    passed = 0
    total = len(results)
    
    for integration_name, result in results:
        status = "✅ 成功" if result else "❌ 失败"
        print(f"{integration_name}: {status}")
        if result:
            passed += 1
    
    print(f"\n总计: {passed}/{total} 集成成功")
    
    if passed == total:
        print("🎉 所有集成完成！FFmpeg服务已成功集成到项目中。")
        print("\n下一步:")
        print("1. 运行环境准备脚本: setup_environment.bat (Windows) 或 setup_environment.sh (Linux/macOS)")
        print("2. 运行构建脚本: build.bat (Windows) 或 build.sh (Linux/macOS)")
        print("3. 运行测试脚本: python test_basic_functionality.py")
        return 0
    else:
        print("⚠️ 部分集成失败，请检查项目结构。")
        return 1

if __name__ == "__main__":
    sys.exit(main()) 