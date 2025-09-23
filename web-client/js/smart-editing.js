/**
 * AI智能剪辑 - Web客户端界面
 * 提供直观的剪辑预览界面，支持手动调整、自定义模板、导出设置等功能
 */

class SmartEditingInterface {
    constructor() {
        this.currentTask = null;
        this.analysisResults = null;
        this.previewPlayer = null;
        this.timeline = null;
        
        this.editingPresets = [
            { 
                id: 'highlight', 
                name: '精彩集锦', 
                duration: 180, 
                description: '提取会议中最精彩的片段',
                icon: '⭐'
            },
            { 
                id: 'summary', 
                name: '会议摘要', 
                duration: 300, 
                description: '生成5分钟会议摘要',
                icon: '📋'
            },
            { 
                id: 'full', 
                name: '完整优化', 
                duration: 0, 
                description: '优化整个会议视频',
                icon: '🎬'
            },
            { 
                id: 'custom', 
                name: '自定义', 
                duration: 600, 
                description: '自定义剪辑设置',
                icon: '⚙️'
            }
        ];
        
        this.init();
    }
    
    init() {
        this.createInterface();
        this.bindEvents();
        this.initializeTimeline();
    }
    
    createInterface() {
        const container = document.getElementById('smart-editing-container');
        if (!container) {
            console.error('Smart editing container not found');
            return;
        }
        
        container.innerHTML = `
            <div class="smart-editing-interface">
                <!-- 标题栏 -->
                <div class="header">
                    <h2>🤖 AI智能剪辑</h2>
                    <div class="header-actions">
                        <button class="btn-help" onclick="smartEditing.showHelp()">❓ 帮助</button>
                    </div>
                </div>
                
                <!-- 主要内容区域 -->
                <div class="main-content">
                    <!-- 左侧控制面板 -->
                    <div class="control-panel">
                        <!-- 视频选择 -->
                        <div class="section video-selection">
                            <h3>📹 选择视频</h3>
                            <div class="file-input-wrapper">
                                <input type="file" id="video-file-input" accept="video/*" style="display: none;">
                                <div class="file-drop-zone" onclick="document.getElementById('video-file-input').click()">
                                    <div class="drop-zone-content">
                                        <span class="drop-icon">📁</span>
                                        <p>点击选择视频文件或拖拽到此处</p>
                                        <small>支持 MP4, AVI, MOV, MKV, WebM 格式</small>
                                    </div>
                                </div>
                            </div>
                            
                            <!-- 视频信息 -->
                            <div class="video-info" style="display: none;">
                                <div class="video-thumbnail">
                                    <canvas id="video-thumbnail-canvas" width="120" height="70"></canvas>
                                </div>
                                <div class="video-details">
                                    <p><strong>时长:</strong> <span id="video-duration">--:--</span></p>
                                    <p><strong>分辨率:</strong> <span id="video-resolution">--</span></p>
                                    <p><strong>大小:</strong> <span id="video-size">--</span></p>
                                </div>
                                <button class="btn-primary" onclick="smartEditing.analyzeVideo()">
                                    🔍 分析视频
                                </button>
                            </div>
                        </div>
                        
                        <!-- 剪辑模板 -->
                        <div class="section preset-selection">
                            <h3>🎨 选择剪辑模板</h3>
                            <div class="preset-grid">
                                ${this.editingPresets.map((preset, index) => `
                                    <div class="preset-card" data-preset="${preset.id}" onclick="smartEditing.selectPreset('${preset.id}')">
                                        <div class="preset-icon">${preset.icon}</div>
                                        <h4>${preset.name}</h4>
                                        <p>${preset.description}</p>
                                        <small>${preset.duration > 0 ? `目标时长: ${Math.floor(preset.duration / 60)}分钟` : '保持原时长'}</small>
                                    </div>
                                `).join('')}
                            </div>
                        </div>
                        
                        <!-- 高级设置 -->
                        <div class="section advanced-settings">
                            <h3>⚙️ 高级设置</h3>
                            <div class="settings-content">
                                <!-- 输出质量 -->
                                <div class="setting-group">
                                    <label>输出质量:</label>
                                    <select id="quality-select">
                                        <option value="high">高质量 (较慢)</option>
                                        <option value="medium" selected>标准质量</option>
                                        <option value="low">快速处理</option>
                                    </select>
                                </div>
                                
                                <!-- 输出格式 -->
                                <div class="setting-group">
                                    <label>输出格式:</label>
                                    <select id="format-select">
                                        <option value="mp4" selected>MP4</option>
                                        <option value="webm">WebM</option>
                                        <option value="avi">AVI</option>
                                    </select>
                                </div>
                                
                                <!-- 特效选项 -->
                                <div class="setting-group">
                                    <label>特效选项:</label>
                                    <div class="checkbox-group">
                                        <label><input type="checkbox" id="add-subtitles" checked> 自动生成字幕</label>
                                        <label><input type="checkbox" id="add-music"> 添加背景音乐</label>
                                        <label><input type="checkbox" id="enhance-audio" checked> 音频增强</label>
                                        <label><input type="checkbox" id="stabilize-video"> 视频防抖</label>
                                    </div>
                                </div>
                                
                                <!-- 滤镜设置 -->
                                <div class="setting-group">
                                    <label>视频滤镜:</label>
                                    <div class="filter-controls">
                                        <div class="filter-control">
                                            <label>亮度: <span id="brightness-value">0</span></label>
                                            <input type="range" id="brightness-slider" min="-50" max="50" value="0" 
                                                   oninput="smartEditing.updateFilterValue('brightness', this.value)">
                                        </div>
                                        <div class="filter-control">
                                            <label>对比度: <span id="contrast-value">1.0</span></label>
                                            <input type="range" id="contrast-slider" min="50" max="200" value="100" 
                                                   oninput="smartEditing.updateFilterValue('contrast', this.value)">
                                        </div>
                                        <div class="filter-control">
                                            <label>饱和度: <span id="saturation-value">1.0</span></label>
                                            <input type="range" id="saturation-slider" min="0" max="200" value="100" 
                                                   oninput="smartEditing.updateFilterValue('saturation', this.value)">
                                        </div>
                                    </div>
                                </div>
                            </div>
                        </div>
                        
                        <!-- 操作按钮 -->
                        <div class="section actions">
                            <button class="btn-secondary" onclick="smartEditing.previewSettings()" disabled>
                                👁️ 预览设置
                            </button>
                            <button class="btn-primary" onclick="smartEditing.startEditing()" disabled>
                                🚀 开始剪辑
                            </button>
                        </div>
                    </div>
                    
                    <!-- 右侧预览区域 -->
                    <div class="preview-panel">
                        <!-- 视频预览 -->
                        <div class="video-preview">
                            <video id="preview-video" controls style="display: none;">
                                您的浏览器不支持视频播放
                            </video>
                            <div class="preview-placeholder">
                                <div class="placeholder-content">
                                    <span class="placeholder-icon">🎬</span>
                                    <p>选择视频文件开始预览</p>
                                </div>
                            </div>
                        </div>
                        
                        <!-- 时间轴 -->
                        <div class="timeline-container">
                            <div class="timeline-header">
                                <h4>📊 智能分析时间轴</h4>
                                <div class="timeline-controls">
                                    <button onclick="smartEditing.zoomTimeline('in')">🔍+</button>
                                    <button onclick="smartEditing.zoomTimeline('out')">🔍-</button>
                                    <button onclick="smartEditing.resetTimeline()">↻</button>
                                </div>
                            </div>
                            <div class="timeline" id="editing-timeline">
                                <div class="timeline-placeholder">
                                    <p>分析视频后将显示智能时间轴</p>
                                </div>
                            </div>
                        </div>
                        
                        <!-- 高光片段列表 -->
                        <div class="highlights-panel">
                            <h4>⭐ 检测到的高光片段</h4>
                            <div class="highlights-list" id="highlights-list">
                                <div class="highlights-placeholder">
                                    <p>分析视频后将显示高光片段</p>
                                </div>
                            </div>
                        </div>
                    </div>
                </div>
                
                <!-- 进度对话框 -->
                <div class="modal" id="progress-modal" style="display: none;">
                    <div class="modal-content">
                        <h3>🤖 AI正在处理中...</h3>
                        <div class="progress-container">
                            <div class="progress-bar">
                                <div class="progress-fill" id="progress-fill"></div>
                            </div>
                            <div class="progress-text">
                                <span id="progress-percentage">0%</span>
                                <span id="progress-status">准备中...</span>
                            </div>
                        </div>
                        <div class="progress-details">
                            <p id="progress-detail">正在初始化处理流程...</p>
                        </div>
                        <button class="btn-secondary" onclick="smartEditing.cancelTask()">取消</button>
                    </div>
                </div>
            </div>
        `;
    }
    
    bindEvents() {
        // 文件选择事件
        const fileInput = document.getElementById('video-file-input');
        if (fileInput) {
            fileInput.addEventListener('change', (e) => this.handleFileSelect(e));
        }
        
        // 拖拽事件
        const dropZone = document.querySelector('.file-drop-zone');
        if (dropZone) {
            dropZone.addEventListener('dragover', (e) => {
                e.preventDefault();
                dropZone.classList.add('drag-over');
            });
            
            dropZone.addEventListener('dragleave', () => {
                dropZone.classList.remove('drag-over');
            });
            
            dropZone.addEventListener('drop', (e) => {
                e.preventDefault();
                dropZone.classList.remove('drag-over');
                const files = e.dataTransfer.files;
                if (files.length > 0) {
                    this.handleFileSelect({ target: { files } });
                }
            });
        }
    }
    
    handleFileSelect(event) {
        const file = event.target.files[0];
        if (!file) return;
        
        // 验证文件类型
        if (!file.type.startsWith('video/')) {
            alert('请选择视频文件');
            return;
        }
        
        // 显示视频信息
        this.displayVideoInfo(file);
        
        // 加载视频到预览器
        this.loadVideoPreview(file);
    }
    
    displayVideoInfo(file) {
        const videoInfo = document.querySelector('.video-info');
        const dropZone = document.querySelector('.file-drop-zone');
        
        if (videoInfo && dropZone) {
            // 隐藏拖拽区域，显示视频信息
            dropZone.style.display = 'none';
            videoInfo.style.display = 'flex';
            
            // 更新文件信息
            document.getElementById('video-size').textContent = this.formatFileSize(file.size);
            
            // 生成缩略图
            this.generateThumbnail(file);
        }
    }
    
    generateThumbnail(file) {
        const video = document.createElement('video');
        const canvas = document.getElementById('video-thumbnail-canvas');
        const ctx = canvas.getContext('2d');
        
        video.addEventListener('loadedmetadata', () => {
            // 更新视频信息
            document.getElementById('video-duration').textContent = this.formatDuration(video.duration);
            document.getElementById('video-resolution').textContent = `${video.videoWidth}x${video.videoHeight}`;
            
            // 跳到视频中间生成缩略图
            video.currentTime = video.duration / 2;
        });
        
        video.addEventListener('seeked', () => {
            // 绘制缩略图
            ctx.drawImage(video, 0, 0, canvas.width, canvas.height);
        });
        
        video.src = URL.createObjectURL(file);
    }
    
    loadVideoPreview(file) {
        const previewVideo = document.getElementById('preview-video');
        const placeholder = document.querySelector('.preview-placeholder');
        
        if (previewVideo && placeholder) {
            previewVideo.src = URL.createObjectURL(file);
            previewVideo.style.display = 'block';
            placeholder.style.display = 'none';
            
            // 启用预览按钮
            document.querySelector('.actions .btn-secondary').disabled = false;
            document.querySelector('.actions .btn-primary').disabled = false;
        }
    }
    
    selectPreset(presetId) {
        // 移除之前的选中状态
        document.querySelectorAll('.preset-card').forEach(card => {
            card.classList.remove('selected');
        });
        
        // 添加选中状态
        const selectedCard = document.querySelector(`[data-preset="${presetId}"]`);
        if (selectedCard) {
            selectedCard.classList.add('selected');
        }
        
        this.selectedPreset = presetId;
        console.log('选择预设:', presetId);
    }
    
    updateFilterValue(filterType, value) {
        let displayValue;
        
        switch (filterType) {
            case 'brightness':
                displayValue = (value / 100).toFixed(1);
                break;
            case 'contrast':
                displayValue = (value / 100).toFixed(1);
                break;
            case 'saturation':
                displayValue = (value / 100).toFixed(1);
                break;
        }
        
        document.getElementById(`${filterType}-value`).textContent = displayValue;
        
        // 实时预览滤镜效果（如果需要）
        this.applyPreviewFilters();
    }
    
    applyPreviewFilters() {
        const video = document.getElementById('preview-video');
        if (!video) return;
        
        const brightness = document.getElementById('brightness-slider').value;
        const contrast = document.getElementById('contrast-slider').value;
        const saturation = document.getElementById('saturation-slider').value;
        
        video.style.filter = `
            brightness(${1 + brightness / 100})
            contrast(${contrast / 100})
            saturate(${saturation / 100})
        `;
    }
    
    async analyzeVideo() {
        console.log('开始分析视频...');
        
        this.showProgressModal('分析中', '正在分析视频内容...');
        
        try {
            // 模拟分析过程
            await this.simulateAnalysis();
            
            // 显示分析结果
            this.displayAnalysisResults();
            
            this.hideProgressModal();
        } catch (error) {
            console.error('视频分析失败:', error);
            alert('视频分析失败，请重试');
            this.hideProgressModal();
        }
    }
    
    async simulateAnalysis() {
        const steps = [
            { progress: 20, status: '提取音频特征...', detail: '正在分析语音活跃度和情绪' },
            { progress: 40, status: '分析视频内容...', detail: '检测人脸表情和动作' },
            { progress: 60, status: '处理文本内容...', detail: '语音转文字和关键词提取' },
            { progress: 80, status: '生成智能评分...', detail: '计算片段重要性评分' },
            { progress: 100, status: '完成分析', detail: '分析完成，生成高光片段' }
        ];
        
        for (const step of steps) {
            await new Promise(resolve => setTimeout(resolve, 1000));
            this.updateProgress(step.progress, step.status, step.detail);
        }
    }
    
    displayAnalysisResults() {
        // 模拟分析结果
        const mockHighlights = [
            { start: 120, end: 180, score: 0.95, type: 'decision', title: '重要决策讨论' },
            { start: 300, end: 350, score: 0.88, type: 'presentation', title: '产品演示' },
            { start: 600, end: 650, score: 0.82, type: 'interaction', title: '激烈讨论' },
            { start: 900, end: 940, score: 0.79, type: 'reaction', title: '情绪反应' }
        ];
        
        this.displayTimeline(mockHighlights);
        this.displayHighlightsList(mockHighlights);
    }
    
    displayTimeline(highlights) {
        const timeline = document.getElementById('editing-timeline');
        const placeholder = timeline.querySelector('.timeline-placeholder');
        
        if (placeholder) {
            placeholder.remove();
        }
        
        // 创建时间轴可视化
        const timelineViz = document.createElement('div');
        timelineViz.className = 'timeline-visualization';
        
        highlights.forEach(highlight => {
            const segment = document.createElement('div');
            segment.className = `timeline-segment ${highlight.type}`;
            segment.style.left = `${(highlight.start / 1800) * 100}%`;
            segment.style.width = `${((highlight.end - highlight.start) / 1800) * 100}%`;
            segment.title = `${highlight.title} (${this.formatDuration(highlight.start)} - ${this.formatDuration(highlight.end)})`;
            
            timelineViz.appendChild(segment);
        });
        
        timeline.appendChild(timelineViz);
    }
    
    displayHighlightsList(highlights) {
        const highlightsList = document.getElementById('highlights-list');
        const placeholder = highlightsList.querySelector('.highlights-placeholder');
        
        if (placeholder) {
            placeholder.remove();
        }
        
        highlights.forEach((highlight, index) => {
            const item = document.createElement('div');
            item.className = 'highlight-item';
            item.innerHTML = `
                <div class="highlight-info">
                    <div class="highlight-title">${highlight.title}</div>
                    <div class="highlight-time">${this.formatDuration(highlight.start)} - ${this.formatDuration(highlight.end)}</div>
                    <div class="highlight-score">评分: ${(highlight.score * 100).toFixed(0)}%</div>
                </div>
                <div class="highlight-actions">
                    <button onclick="smartEditing.previewHighlight(${index})">预览</button>
                    <button onclick="smartEditing.toggleHighlight(${index})">包含</button>
                </div>
            `;
            
            highlightsList.appendChild(item);
        });
    }
    
    startEditing() {
        console.log('开始智能剪辑...');
        
        const config = this.getEditingConfig();
        console.log('剪辑配置:', config);
        
        this.showProgressModal('剪辑中', '正在生成智能剪辑...');
        
        // 模拟剪辑过程
        this.simulateEditing();
    }
    
    async simulateEditing() {
        const steps = [
            { progress: 15, status: '准备素材...', detail: '加载视频文件和分析结果' },
            { progress: 30, status: '提取片段...', detail: '根据评分提取高光片段' },
            { progress: 50, status: '应用滤镜...', detail: '处理视频滤镜和特效' },
            { progress: 70, status: '生成转场...', detail: '添加转场效果和音频处理' },
            { progress: 85, status: '渲染视频...', detail: '最终渲染和编码' },
            { progress: 100, status: '完成', detail: '智能剪辑完成！' }
        ];
        
        for (const step of steps) {
            await new Promise(resolve => setTimeout(resolve, 2000));
            this.updateProgress(step.progress, step.status, step.detail);
        }
        
        // 显示完成对话框
        setTimeout(() => {
            this.hideProgressModal();
            this.showCompletionDialog();
        }, 1000);
    }
    
    getEditingConfig() {
        return {
            preset: this.selectedPreset || 'highlight',
            quality: document.getElementById('quality-select').value,
            format: document.getElementById('format-select').value,
            addSubtitles: document.getElementById('add-subtitles').checked,
            addMusic: document.getElementById('add-music').checked,
            enhanceAudio: document.getElementById('enhance-audio').checked,
            stabilizeVideo: document.getElementById('stabilize-video').checked,
            filters: {
                brightness: document.getElementById('brightness-slider').value / 100,
                contrast: document.getElementById('contrast-slider').value / 100,
                saturation: document.getElementById('saturation-slider').value / 100
            }
        };
    }
    
    showProgressModal(title, status) {
        const modal = document.getElementById('progress-modal');
        const titleEl = modal.querySelector('h3');
        const statusEl = document.getElementById('progress-status');
        
        titleEl.textContent = `🤖 ${title}`;
        statusEl.textContent = status;
        modal.style.display = 'flex';
    }
    
    hideProgressModal() {
        document.getElementById('progress-modal').style.display = 'none';
    }
    
    updateProgress(percentage, status, detail) {
        document.getElementById('progress-fill').style.width = `${percentage}%`;
        document.getElementById('progress-percentage').textContent = `${percentage}%`;
        document.getElementById('progress-status').textContent = status;
        document.getElementById('progress-detail').textContent = detail;
    }
    
    showCompletionDialog() {
        alert('🎉 智能剪辑完成！\n\n您的视频已经成功处理，可以在下载区域获取结果。');
    }
    
    // 工具函数
    formatFileSize(bytes) {
        const sizes = ['B', 'KB', 'MB', 'GB'];
        if (bytes === 0) return '0 B';
        const i = Math.floor(Math.log(bytes) / Math.log(1024));
        return Math.round(bytes / Math.pow(1024, i) * 100) / 100 + ' ' + sizes[i];
    }
    
    formatDuration(seconds) {
        const hours = Math.floor(seconds / 3600);
        const minutes = Math.floor((seconds % 3600) / 60);
        const secs = Math.floor(seconds % 60);
        
        if (hours > 0) {
            return `${hours}:${minutes.toString().padStart(2, '0')}:${secs.toString().padStart(2, '0')}`;
        } else {
            return `${minutes}:${secs.toString().padStart(2, '0')}`;
        }
    }
    
    showHelp() {
        alert(`🤖 AI智能剪辑帮助

1. 精彩集锦模式
   - 自动识别会议中的重要时刻
   - 提取最有价值的片段
   - 适合快速回顾会议要点

2. 会议摘要模式
   - 生成5分钟左右的会议摘要
   - 包含关键决策和讨论
   - 适合分享给未参会人员

3. 完整优化模式
   - 保持完整会议内容
   - 优化音视频质量
   - 添加字幕和标记

4. 自定义模式
   - 可自定义时长和内容
   - 灵活的剪辑选项
   - 适合特殊需求

使用建议：
- 确保视频文件完整且清晰
- 选择合适的输出质量
- 根据用途选择合适的模板`);
    }
}

// 初始化智能剪辑界面
let smartEditing;
document.addEventListener('DOMContentLoaded', () => {
    smartEditing = new SmartEditingInterface();
});
