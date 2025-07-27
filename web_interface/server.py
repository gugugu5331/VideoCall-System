#!/usr/bin/env python3
# -*- coding: utf-8 -*-

import http.server
import socketserver
import os
import sys
import webbrowser
from urllib.parse import urlparse

class CORSHTTPRequestHandler(http.server.SimpleHTTPRequestHandler):
    def end_headers(self):
        self.send_header('Access-Control-Allow-Origin', '*')
        self.send_header('Access-Control-Allow-Methods', 'GET, POST, OPTIONS')
        self.send_header('Access-Control-Allow-Headers', 'Content-Type, Authorization')
        super().end_headers()
    
    def do_OPTIONS(self):
        self.send_response(200)
        self.end_headers()

def main():
    # 设置端口
    PORT = 8080
    
    # 切换到web_interface目录
    script_dir = os.path.dirname(os.path.abspath(__file__))
    os.chdir(script_dir)
    
    # 检查文件是否存在
    if not os.path.exists('index.html'):
        print("错误: 找不到 index.html 文件")
        print(f"当前目录: {os.getcwd()}")
        print(f"文件列表: {os.listdir('.')}")
        return
    
    # 创建服务器
    with socketserver.TCPServer(("", PORT), CORSHTTPRequestHandler) as httpd:
        print("=" * 50)
        print("🌐 音视频通话系统 - Web界面")
        print("=" * 50)
        print(f"📁 服务目录: {os.getcwd()}")
        print(f"🌍 访问地址: http://localhost:{PORT}")
        print(f"🔗 后端API: http://localhost:8000")
        print(f"🤖 AI服务: http://localhost:5001")
        print("=" * 50)
        print("💡 提示:")
        print("   - 确保后端服务正在运行")
        print("   - 确保AI服务正在运行")
        print("   - 浏览器会自动打开Web界面")
        print("   - 按 Ctrl+C 停止服务器")
        print("=" * 50)
        
        # 自动打开浏览器
        try:
            webbrowser.open(f'http://localhost:{PORT}')
            print("✅ 浏览器已自动打开")
        except:
            print("⚠️  无法自动打开浏览器，请手动访问")
        
        print("\n🚀 服务器启动中...")
        
        try:
            httpd.serve_forever()
        except KeyboardInterrupt:
            print("\n\n🛑 服务器已停止")
            print("感谢使用音视频通话系统！")

if __name__ == "__main__":
    main() 