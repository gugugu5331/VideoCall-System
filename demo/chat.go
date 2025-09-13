package main

import (
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"
	"unicode/utf8"
)

// ChatMessage 聊天消息结构
type ChatMessage struct {
	ID        string    `json:"id"`
	Type      string    `json:"type"`
	Content   string    `json:"content"`
	Sender    string    `json:"sender"`
	SenderID  string    `json:"senderId"`
	MeetingID string    `json:"meetingId"`
	Timestamp time.Time `json:"timestamp"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
}

// ChatManager 聊天管理器
type ChatManager struct {
	hub      *SignalingHub
	messages map[string][]*ChatMessage // meetingId -> messages
	maxMessages int // 每个会议最大保存的消息数量
}

// NewChatManager 创建新的聊天管理器
func NewChatManager(hub *SignalingHub) *ChatManager {
	return &ChatManager{
		hub:         hub,
		messages:    make(map[string][]*ChatMessage),
		maxMessages: 100, // 每个会议最多保存100条消息
	}
}

// HandleChatMessage 处理聊天消息
func (cm *ChatManager) HandleChatMessage(client *WebSocketClient, rawMessage map[string]interface{}) error {
	// 验证客户端状态
	if err := cm.validateClient(client); err != nil {
		log.Printf("聊天消息验证失败: %v", err)
		return err
	}

	// 解析消息数据
	data, ok := rawMessage["data"].(map[string]interface{})
	if !ok {
		return fmt.Errorf("无效的消息数据格式")
	}

	// 创建聊天消息
	chatMsg, err := cm.createChatMessage(client, data)
	if err != nil {
		log.Printf("创建聊天消息失败: %v", err)
		return err
	}

	// 验证消息内容
	if err := cm.validateMessage(chatMsg); err != nil {
		log.Printf("消息内容验证失败: %v", err)
		return err
	}

	// 保存消息到历史记录
	cm.saveMessage(chatMsg)

	// 广播消息
	if err := cm.broadcastMessage(chatMsg); err != nil {
		log.Printf("广播消息失败: %v", err)
		return err
	}

	log.Printf("聊天消息处理成功: %s -> %s (会议: %s)", 
		chatMsg.Sender, chatMsg.Content, chatMsg.MeetingID)

	return nil
}

// validateClient 验证客户端状态
func (cm *ChatManager) validateClient(client *WebSocketClient) error {
	if client == nil {
		return fmt.Errorf("客户端为空")
	}

	if client.MeetingID == "" {
		return fmt.Errorf("客户端未加入会议")
	}

	if client.Username == "" {
		return fmt.Errorf("客户端用户名为空")
	}

	if client.UserID == "" {
		return fmt.Errorf("客户端用户ID为空")
	}

	// 检查客户端是否在会议中
	if meetingClients, exists := cm.hub.meetings[client.MeetingID]; exists {
		if _, inMeeting := meetingClients[client]; !inMeeting {
			return fmt.Errorf("客户端不在指定会议中")
		}
	} else {
		return fmt.Errorf("会议不存在: %s", client.MeetingID)
	}

	return nil
}

// createChatMessage 创建聊天消息
func (cm *ChatManager) createChatMessage(client *WebSocketClient, data map[string]interface{}) (*ChatMessage, error) {
	// 提取消息内容
	content, ok := data["message"].(string)
	if !ok {
		return nil, fmt.Errorf("消息内容不是字符串类型")
	}

	// 生成消息ID
	messageID := fmt.Sprintf("%s_%d_%s", client.MeetingID, time.Now().UnixNano(), client.UserID)

	// 提取消息类型（默认为text）
	msgType := "text"
	if t, ok := data["type"].(string); ok {
		msgType = t
	}

	// 提取元数据
	metadata := make(map[string]interface{})
	if meta, ok := data["metadata"].(map[string]interface{}); ok {
		metadata = meta
	}

	chatMsg := &ChatMessage{
		ID:        messageID,
		Type:      msgType,
		Content:   content,
		Sender:    client.Username,
		SenderID:  client.UserID,
		MeetingID: client.MeetingID,
		Timestamp: time.Now(),
		Metadata:  metadata,
	}

	return chatMsg, nil
}

// validateMessage 验证消息内容
func (cm *ChatManager) validateMessage(msg *ChatMessage) error {
	// 检查消息内容长度
	if len(strings.TrimSpace(msg.Content)) == 0 {
		return fmt.Errorf("消息内容不能为空")
	}

	// 检查消息长度（UTF-8字符数）
	if utf8.RuneCountInString(msg.Content) > 1000 {
		return fmt.Errorf("消息内容过长，最大1000个字符")
	}

	// 检查消息类型
	validTypes := map[string]bool{
		"text":  true,
		"emoji": true,
		"file":  true,
		"image": true,
	}
	if !validTypes[msg.Type] {
		return fmt.Errorf("不支持的消息类型: %s", msg.Type)
	}

	// 基本的内容过滤（可以扩展为更复杂的过滤规则）
	if cm.containsInappropriateContent(msg.Content) {
		return fmt.Errorf("消息包含不当内容")
	}

	return nil
}

// containsInappropriateContent 检查是否包含不当内容
func (cm *ChatManager) containsInappropriateContent(content string) bool {
	// 这里可以实现更复杂的内容过滤逻辑
	// 例如：敏感词过滤、垃圾信息检测等
	
	// 简单的示例：检查是否包含某些关键词
	inappropriateWords := []string{
		// 可以添加需要过滤的词汇
	}

	contentLower := strings.ToLower(content)
	for _, word := range inappropriateWords {
		if strings.Contains(contentLower, word) {
			return true
		}
	}

	return false
}

// saveMessage 保存消息到历史记录
func (cm *ChatManager) saveMessage(msg *ChatMessage) {
	// 获取会议的消息历史
	messages, exists := cm.messages[msg.MeetingID]
	if !exists {
		messages = make([]*ChatMessage, 0)
	}

	// 添加新消息
	messages = append(messages, msg)

	// 如果消息数量超过限制，删除最旧的消息
	if len(messages) > cm.maxMessages {
		messages = messages[len(messages)-cm.maxMessages:]
	}

	// 保存回映射
	cm.messages[msg.MeetingID] = messages

	log.Printf("消息已保存: 会议 %s 现有 %d 条消息", msg.MeetingID, len(messages))
}

// broadcastMessage 广播消息到会议中的所有用户
func (cm *ChatManager) broadcastMessage(msg *ChatMessage) error {
	// 构建广播消息
	broadcastData := map[string]interface{}{
		"type": "chat-message",
		"data": map[string]interface{}{
			"id":        msg.ID,
			"type":      msg.Type,
			"message":   msg.Content,
			"sender":    msg.Sender,
			"senderId":  msg.SenderID,
			"timestamp": msg.Timestamp.Format("2006-01-02 15:04:05"),
			"metadata":  msg.Metadata,
		},
	}

	// 获取会议中的所有客户端
	meetingClients, exists := cm.hub.meetings[msg.MeetingID]
	if !exists {
		return fmt.Errorf("会议不存在: %s", msg.MeetingID)
	}

	log.Printf("开始广播聊天消息到会议 %s，共 %d 个用户", msg.MeetingID, len(meetingClients))

	// 统计发送结果
	successCount := 0
	failureCount := 0

	// 向每个客户端发送消息
	for client := range meetingClients {
		if cm.sendMessageToClient(client, broadcastData) {
			successCount++
			log.Printf("消息发送成功: %s(%s)", client.UserID, client.Username)
		} else {
			failureCount++
			log.Printf("消息发送失败: %s(%s)", client.UserID, client.Username)
		}
	}

	log.Printf("消息广播完成: 成功 %d，失败 %d", successCount, failureCount)

	// 如果所有发送都失败，返回错误
	if successCount == 0 && failureCount > 0 {
		return fmt.Errorf("消息广播失败，所有客户端都无法接收")
	}

	return nil
}

// sendMessageToClient 向单个客户端发送消息
func (cm *ChatManager) sendMessageToClient(client *WebSocketClient, message map[string]interface{}) bool {
	if client == nil || client.Send == nil {
		return false
	}

	select {
	case client.Send <- message:
		return true
	case <-time.After(100 * time.Millisecond):
		// 发送超时
		return false
	}
}

// GetChatHistory 获取会议的聊天历史
func (cm *ChatManager) GetChatHistory(meetingID string, limit int) []*ChatMessage {
	messages, exists := cm.messages[meetingID]
	if !exists {
		return []*ChatMessage{}
	}

	// 如果请求的数量超过现有消息数量，返回所有消息
	if limit <= 0 || limit > len(messages) {
		limit = len(messages)
	}

	// 返回最新的limit条消息
	start := len(messages) - limit
	if start < 0 {
		start = 0
	}

	return messages[start:]
}

// ClearChatHistory 清除会议的聊天历史
func (cm *ChatManager) ClearChatHistory(meetingID string) {
	delete(cm.messages, meetingID)
	log.Printf("已清除会议 %s 的聊天历史", meetingID)
}

// GetChatStats 获取聊天统计信息
func (cm *ChatManager) GetChatStats(meetingID string) map[string]interface{} {
	messages, exists := cm.messages[meetingID]
	if !exists {
		return map[string]interface{}{
			"totalMessages": 0,
			"participants":  0,
		}
	}

	// 统计参与者
	participants := make(map[string]bool)
	for _, msg := range messages {
		participants[msg.SenderID] = true
	}

	return map[string]interface{}{
		"totalMessages": len(messages),
		"participants":  len(participants),
		"lastMessage":   messages[len(messages)-1].Timestamp,
	}
}

// ExportChatHistory 导出聊天历史（JSON格式）
func (cm *ChatManager) ExportChatHistory(meetingID string) ([]byte, error) {
	messages := cm.GetChatHistory(meetingID, 0) // 获取所有消息
	
	exportData := map[string]interface{}{
		"meetingId":   meetingID,
		"exportTime":  time.Now(),
		"totalCount":  len(messages),
		"messages":    messages,
	}

	return json.MarshalIndent(exportData, "", "  ")
}
