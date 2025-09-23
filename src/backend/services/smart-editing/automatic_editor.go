package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

// AutomaticEditor 自动剪辑器
type AutomaticEditor struct {
	storagePath string
	tempPath    string
}

// NewAutomaticEditor 创建自动剪辑器
func NewAutomaticEditor(storagePath string) (*AutomaticEditor, error) {
	tempPath := filepath.Join(storagePath, "temp")
	
	// 创建必要的目录
	if err := os.MkdirAll(storagePath, 0755); err != nil {
		return nil, fmt.Errorf("failed to create storage directory: %w", err)
	}
	
	if err := os.MkdirAll(tempPath, 0755); err != nil {
		return nil, fmt.Errorf("failed to create temp directory: %w", err)
	}
	
	return &AutomaticEditor{
		storagePath: storagePath,
		tempPath:    tempPath,
	}, nil
}

// CreateSmartEdit 创建智能剪辑
func (e *AutomaticEditor) CreateSmartEdit(videoPath string, analysis ContentAnalysis, config EditingConfig) (*EditingResult, error) {
	log.Printf("开始智能剪辑: %s", videoPath)
	
	// 1. 提取高光片段
	highlights := analysis.Highlights
	if len(highlights) == 0 {
		return nil, fmt.Errorf("no highlights found for editing")
	}
	
	// 2. 生成输出文件名
	outputPath := e.generateOutputPath(config.Format)
	
	// 3. 根据剪辑风格处理
	var result *EditingResult
	var err error
	
	switch config.Style {
	case "highlight":
		result, err = e.createHighlightReel(videoPath, highlights, outputPath, config)
	case "summary":
		result, err = e.createSummaryVideo(videoPath, analysis, outputPath, config)
	case "full":
		result, err = e.createFullEditedVideo(videoPath, analysis, outputPath, config)
	default:
		result, err = e.createCustomEdit(videoPath, analysis, outputPath, config)
	}
	
	if err != nil {
		return nil, fmt.Errorf("failed to create edit: %w", err)
	}
	
	// 4. 后处理
	if err := e.postProcess(result, config); err != nil {
		log.Printf("Post-processing warning: %v", err)
	}
	
	log.Printf("智能剪辑完成: %s", result.OutputPath)
	return result, nil
}

// createHighlightReel 创建高光集锦
func (e *AutomaticEditor) createHighlightReel(videoPath string, highlights []Highlight, outputPath string, config EditingConfig) (*EditingResult, error) {
	log.Printf("创建高光集锦，包含 %d 个片段", len(highlights))
	
	// 生成FFmpeg过滤器
	var filterParts []string
	var segments []EditedSegment
	
	currentTime := 0.0
	
	for i, highlight := range highlights {
		duration := highlight.EndTime - highlight.StartTime
		
		// 限制单个片段最大时长
		if duration > 30.0 {
			duration = 30.0
		}
		
		// 添加片段到过滤器
		filterParts = append(filterParts, fmt.Sprintf(
			"[0:v]trim=start=%.2f:end=%.2f,setpts=PTS-STARTPTS[v%d]; [0:a]atrim=start=%.2f:end=%.2f,asetpts=PTS-STARTPTS[a%d]",
			highlight.StartTime, highlight.StartTime+duration, i,
			highlight.StartTime, highlight.StartTime+duration, i,
		))
		
		segments = append(segments, EditedSegment{
			OriginalStart: highlight.StartTime,
			OriginalEnd:   highlight.StartTime + duration,
			EditedStart:   currentTime,
			EditedEnd:     currentTime + duration,
			Type:          highlight.Type,
			Importance:    highlight.Score,
		})
		
		currentTime += duration
	}
	
	// 连接所有片段
	var concatInputs []string
	for i := 0; i < len(highlights); i++ {
		concatInputs = append(concatInputs, fmt.Sprintf("[v%d][a%d]", i, i))
	}
	
	filterComplex := strings.Join(filterParts, "; ") + "; " +
		strings.Join(concatInputs, "") + fmt.Sprintf("concat=n=%d:v=1:a=1[outv][outa]", len(highlights))
	
	// 构建FFmpeg命令
	args := []string{
		"-i", videoPath,
		"-filter_complex", filterComplex,
		"-map", "[outv]",
		"-map", "[outa]",
		"-c:v", "libx264",
		"-preset", "medium",
		"-crf", "23",
		"-c:a", "aac",
		"-b:a", "128k",
		"-y", outputPath,
	}
	
	if err := e.runFFmpeg(args); err != nil {
		return nil, fmt.Errorf("failed to create highlight reel: %w", err)
	}
	
	// 生成转场效果
	transitions := e.generateTransitions(segments, config)
	
	// 获取输出文件信息
	fileInfo, err := os.Stat(outputPath)
	if err != nil {
		return nil, fmt.Errorf("failed to get output file info: %w", err)
	}
	
	result := &EditingResult{
		OutputPath:  outputPath,
		Duration:    currentTime,
		FileSize:    fileInfo.Size(),
		Resolution:  "1280x720", // 默认分辨率
		Bitrate:     2000,       // 默认比特率
		Segments:    segments,
		Transitions: transitions,
		Effects:     []Effect{},
		Subtitles:   []Subtitle{},
	}
	
	return result, nil
}

// createSummaryVideo 创建摘要视频
func (e *AutomaticEditor) createSummaryVideo(videoPath string, analysis ContentAnalysis, outputPath string, config EditingConfig) (*EditingResult, error) {
	log.Printf("创建摘要视频，目标时长: %d 秒", config.Duration)
	
	// 选择最重要的片段，总时长不超过目标时长
	selectedHighlights := e.selectHighlightsForDuration(analysis.Highlights, float64(config.Duration))
	
	// 使用高光集锦的方法，但添加更多的转场效果
	result, err := e.createHighlightReel(videoPath, selectedHighlights, outputPath, config)
	if err != nil {
		return nil, err
	}
	
	// 为摘要视频添加开场和结尾
	if err := e.addIntroAndOutro(result, config); err != nil {
		log.Printf("Failed to add intro/outro: %v", err)
	}
	
	return result, nil
}

// createFullEditedVideo 创建完整编辑视频
func (e *AutomaticEditor) createFullEditedVideo(videoPath string, analysis ContentAnalysis, outputPath string, config EditingConfig) (*EditingResult, error) {
	log.Printf("创建完整编辑视频")
	
	// 对整个视频进行优化，但保持完整性
	args := []string{
		"-i", videoPath,
		"-vf", "scale=1280:720,fps=30", // 标准化分辨率和帧率
		"-c:v", "libx264",
		"-preset", "medium",
		"-crf", "23",
		"-c:a", "aac",
		"-b:a", "128k",
		"-y", outputPath,
	}
	
	// 添加视频滤镜
	if len(config.Filters) > 0 {
		filterChain := e.buildFilterChain(config.Filters)
		args = e.insertFilter(args, filterChain)
	}
	
	if err := e.runFFmpeg(args); err != nil {
		return nil, fmt.Errorf("failed to create full edited video: %w", err)
	}
	
	// 获取输出文件信息
	fileInfo, err := os.Stat(outputPath)
	if err != nil {
		return nil, fmt.Errorf("failed to get output file info: %w", err)
	}
	
	result := &EditingResult{
		OutputPath: outputPath,
		Duration:   analysis.Duration,
		FileSize:   fileInfo.Size(),
		Resolution: "1280x720",
		Bitrate:    2000,
		Segments:   []EditedSegment{},
		Effects:    []Effect{},
		Subtitles:  []Subtitle{},
	}
	
	return result, nil
}

// createCustomEdit 创建自定义剪辑
func (e *AutomaticEditor) createCustomEdit(videoPath string, analysis ContentAnalysis, outputPath string, config EditingConfig) (*EditingResult, error) {
	log.Printf("创建自定义剪辑")
	
	// 根据自定义设置处理
	customDuration, ok := config.CustomSettings["duration"].(float64)
	if !ok {
		customDuration = 300.0 // 默认5分钟
	}
	
	selectedHighlights := e.selectHighlightsForDuration(analysis.Highlights, customDuration)
	
	return e.createHighlightReel(videoPath, selectedHighlights, outputPath, config)
}

// selectHighlightsForDuration 选择指定时长的高光片段
func (e *AutomaticEditor) selectHighlightsForDuration(highlights []Highlight, targetDuration float64) []Highlight {
	if len(highlights) == 0 {
		return highlights
	}
	
	// 按评分排序
	sortedHighlights := make([]Highlight, len(highlights))
	copy(sortedHighlights, highlights)
	
	// 简单的冒泡排序（按评分降序）
	for i := 0; i < len(sortedHighlights)-1; i++ {
		for j := 0; j < len(sortedHighlights)-i-1; j++ {
			if sortedHighlights[j].Score < sortedHighlights[j+1].Score {
				sortedHighlights[j], sortedHighlights[j+1] = sortedHighlights[j+1], sortedHighlights[j]
			}
		}
	}
	
	// 选择片段直到达到目标时长
	var selected []Highlight
	totalDuration := 0.0
	
	for _, highlight := range sortedHighlights {
		segmentDuration := highlight.EndTime - highlight.StartTime
		if segmentDuration > 30.0 {
			segmentDuration = 30.0 // 限制单个片段最大时长
		}
		
		if totalDuration+segmentDuration <= targetDuration {
			selected = append(selected, highlight)
			totalDuration += segmentDuration
		}
		
		if totalDuration >= targetDuration*0.9 { // 达到目标时长的90%即可
			break
		}
	}
	
	return selected
}

// generateTransitions 生成转场效果
func (e *AutomaticEditor) generateTransitions(segments []EditedSegment, config EditingConfig) []Transition {
	var transitions []Transition
	
	for i := 0; i < len(segments)-1; i++ {
		transition := Transition{
			Position: segments[i].EditedEnd,
			Type:     "fade",
			Duration: 0.5,
		}
		transitions = append(transitions, transition)
	}
	
	return transitions
}

// buildFilterChain 构建滤镜链
func (e *AutomaticEditor) buildFilterChain(filters []FilterConfig) string {
	var filterParts []string
	
	for _, filter := range filters {
		if !filter.Enabled {
			continue
		}
		
		switch filter.Type {
		case "beauty":
			filterParts = append(filterParts, fmt.Sprintf("gblur=sigma=%.2f", filter.Intensity))
		case "brightness":
			filterParts = append(filterParts, fmt.Sprintf("eq=brightness=%.2f", filter.Intensity-0.5))
		case "contrast":
			filterParts = append(filterParts, fmt.Sprintf("eq=contrast=%.2f", filter.Intensity))
		case "saturation":
			filterParts = append(filterParts, fmt.Sprintf("eq=saturation=%.2f", filter.Intensity))
		}
	}
	
	return strings.Join(filterParts, ",")
}

// insertFilter 在FFmpeg参数中插入滤镜
func (e *AutomaticEditor) insertFilter(args []string, filterChain string) []string {
	if filterChain == "" {
		return args
	}
	
	// 找到-vf参数的位置
	for i, arg := range args {
		if arg == "-vf" && i+1 < len(args) {
			args[i+1] = args[i+1] + "," + filterChain
			return args
		}
	}
	
	// 如果没有找到-vf参数，添加新的
	newArgs := make([]string, 0, len(args)+2)
	inserted := false
	
	for i, arg := range args {
		newArgs = append(newArgs, arg)
		if arg == "-i" && i+1 < len(args) {
			newArgs = append(newArgs, args[i+1])
			if !inserted {
				newArgs = append(newArgs, "-vf", filterChain)
				inserted = true
			}
			i++ // 跳过下一个参数
		}
	}
	
	return newArgs
}

// postProcess 后处理
func (e *AutomaticEditor) postProcess(result *EditingResult, config EditingConfig) error {
	// 生成缩略图
	if err := e.generateThumbnail(result); err != nil {
		log.Printf("Failed to generate thumbnail: %v", err)
	}
	
	// 生成预览视频
	if err := e.generatePreview(result); err != nil {
		log.Printf("Failed to generate preview: %v", err)
	}
	
	// 添加字幕
	if config.AddSubtitles {
		if err := e.addSubtitles(result, config); err != nil {
			log.Printf("Failed to add subtitles: %v", err)
		}
	}
	
	// 添加背景音乐
	if config.AddMusic {
		if err := e.addBackgroundMusic(result, config); err != nil {
			log.Printf("Failed to add background music: %v", err)
		}
	}
	
	return nil
}

// generateOutputPath 生成输出文件路径
func (e *AutomaticEditor) generateOutputPath(format string) string {
	timestamp := time.Now().Format("20060102_150405")
	filename := fmt.Sprintf("edited_%s.%s", timestamp, format)
	return filepath.Join(e.storagePath, filename)
}

// runFFmpeg 运行FFmpeg命令
func (e *AutomaticEditor) runFFmpeg(args []string) error {
	cmd := exec.Command("ffmpeg", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	
	log.Printf("Running FFmpeg: %s", strings.Join(args, " "))
	
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("ffmpeg command failed: %w", err)
	}
	
	return nil
}

// generateThumbnail 生成缩略图
func (e *AutomaticEditor) generateThumbnail(result *EditingResult) error {
	thumbnailPath := strings.TrimSuffix(result.OutputPath, filepath.Ext(result.OutputPath)) + "_thumb.jpg"
	
	args := []string{
		"-i", result.OutputPath,
		"-ss", "00:00:01",
		"-vframes", "1",
		"-q:v", "2",
		"-y", thumbnailPath,
	}
	
	if err := e.runFFmpeg(args); err != nil {
		return err
	}
	
	result.Thumbnail = thumbnailPath
	return nil
}

// generatePreview 生成预览视频
func (e *AutomaticEditor) generatePreview(result *EditingResult) error {
	previewPath := strings.TrimSuffix(result.OutputPath, filepath.Ext(result.OutputPath)) + "_preview.mp4"
	
	args := []string{
		"-i", result.OutputPath,
		"-t", "30", // 30秒预览
		"-vf", "scale=640:360",
		"-c:v", "libx264",
		"-preset", "fast",
		"-crf", "28",
		"-c:a", "aac",
		"-b:a", "64k",
		"-y", previewPath,
	}
	
	if err := e.runFFmpeg(args); err != nil {
		return err
	}
	
	result.Preview = previewPath
	return nil
}

// addSubtitles 添加字幕
func (e *AutomaticEditor) addSubtitles(result *EditingResult, config EditingConfig) error {
	// 这里应该集成语音识别服务生成字幕
	// 暂时跳过实现
	log.Printf("Subtitle generation not implemented yet")
	return nil
}

// addBackgroundMusic 添加背景音乐
func (e *AutomaticEditor) addBackgroundMusic(result *EditingResult, config EditingConfig) error {
	// 这里应该添加背景音乐混合逻辑
	// 暂时跳过实现
	log.Printf("Background music not implemented yet")
	return nil
}

// addIntroAndOutro 添加开场和结尾
func (e *AutomaticEditor) addIntroAndOutro(result *EditingResult, config EditingConfig) error {
	// 这里应该添加开场和结尾片段
	// 暂时跳过实现
	log.Printf("Intro/outro not implemented yet")
	return nil
}
