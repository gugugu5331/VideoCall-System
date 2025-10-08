package database

import (
	"context"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"meeting-system/shared/config"
	"meeting-system/shared/logger"
)

var MongoDB *mongo.Database

// InitMongoDB 初始化MongoDB连接
func InitMongoDB(config config.MongoConfig) error {
	connectTimeout := time.Duration(config.Timeout) * time.Second
	if connectTimeout <= 0 {
		connectTimeout = 30 * time.Second
	}

	ctx, cancel := context.WithTimeout(context.Background(), connectTimeout)
	defer cancel()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(config.URI))
	if err != nil {
		return fmt.Errorf("failed to connect to MongoDB: %w", err)
	}

	if err := client.Ping(ctx, nil); err != nil {
		disconnectCtx, disconnectCancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer disconnectCancel()
		_ = client.Disconnect(disconnectCtx)
		return fmt.Errorf("failed to ping MongoDB: %w", err)
	}

	database := client.Database(config.Database)

	indexCtx, indexCancel := context.WithTimeout(context.Background(), connectTimeout)
	defer indexCancel()

	if err := createMongoIndexes(indexCtx, database); err != nil {
		disconnectCtx, disconnectCancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer disconnectCancel()
		_ = client.Disconnect(disconnectCtx)
		return fmt.Errorf("failed to create MongoDB indexes: %w", err)
	}

	MongoDB = database
	logger.Info("MongoDB connected successfully")

	return nil
}

// GetMongoDB 获取MongoDB数据库实例
func GetMongoDB() *mongo.Database {
	return MongoDB
}

// CloseMongoDB 关闭MongoDB连接
func CloseMongoDB() error {
	if MongoDB == nil {
		return nil
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	client := MongoDB.Client()
	MongoDB = nil
	return client.Disconnect(ctx)
}

// createMongoIndexes 创建MongoDB索引
func createMongoIndexes(ctx context.Context, db *mongo.Database) error {
	// 聊天消息集合索引
	chatCollection := db.Collection("chat_messages")
	chatIndexes := []mongo.IndexModel{
		{
			Keys: map[string]interface{}{
				"meeting_id": 1,
				"timestamp":  -1,
			},
		},
		{
			Keys: map[string]interface{}{
				"user_id": 1,
			},
		},
		{
			Keys: map[string]interface{}{
				"message_type": 1,
			},
		},
	}
	if _, err := chatCollection.Indexes().CreateMany(ctx, chatIndexes); err != nil {
		return fmt.Errorf("failed to create chat_messages indexes: %w", err)
	}

	// AI分析结果集合索引
	aiCollection := db.Collection("ai_analysis_results")
	aiIndexes := []mongo.IndexModel{
		{
			Keys: map[string]interface{}{
				"meeting_id": 1,
				"timestamp":  -1,
			},
		},
		{
			Keys: map[string]interface{}{
				"analysis_type": 1,
			},
		},
		{
			Keys: map[string]interface{}{
				"user_id": 1,
			},
		},
	}
	if _, err := aiCollection.Indexes().CreateMany(ctx, aiIndexes); err != nil {
		return fmt.Errorf("failed to create ai_analysis_results indexes: %w", err)
	}

	// 会议事件集合索引
	eventCollection := db.Collection("meeting_events")
	eventIndexes := []mongo.IndexModel{
		{
			Keys: map[string]interface{}{
				"meeting_id": 1,
				"timestamp":  -1,
			},
		},
		{
			Keys: map[string]interface{}{
				"event_type": 1,
			},
		},
		{
			Keys: map[string]interface{}{
				"user_id": 1,
			},
		},
	}
	if _, err := eventCollection.Indexes().CreateMany(ctx, eventIndexes); err != nil {
		return fmt.Errorf("failed to create meeting_events indexes: %w", err)
	}

	logger.Info("MongoDB indexes created successfully")
	return nil
}

// ChatMessage 聊天消息文档结构
type ChatMessage struct {
	ID          string                 `bson:"_id,omitempty" json:"id"`
	MeetingID   string                 `bson:"meeting_id" json:"meeting_id"`
	UserID      string                 `bson:"user_id" json:"user_id"`
	Username    string                 `bson:"username" json:"username"`
	MessageType string                 `bson:"message_type" json:"message_type"` // text, file, image, emoji
	Content     string                 `bson:"content" json:"content"`
	FileInfo    map[string]interface{} `bson:"file_info,omitempty" json:"file_info,omitempty"`
	Timestamp   time.Time              `bson:"timestamp" json:"timestamp"`
	Metadata    map[string]interface{} `bson:"metadata,omitempty" json:"metadata,omitempty"`
}

// AIAnalysisResult AI分析结果文档结构
type AIAnalysisResult struct {
	ID           string                 `bson:"_id,omitempty" json:"id"`
	MeetingID    string                 `bson:"meeting_id" json:"meeting_id"`
	UserID       string                 `bson:"user_id,omitempty" json:"user_id,omitempty"`
	AnalysisType string                 `bson:"analysis_type" json:"analysis_type"` // emotion, speech, gesture, quality
	Result       map[string]interface{} `bson:"result" json:"result"`
	Confidence   float64                `bson:"confidence" json:"confidence"`
	Timestamp    time.Time              `bson:"timestamp" json:"timestamp"`
	Metadata     map[string]interface{} `bson:"metadata,omitempty" json:"metadata,omitempty"`
}

// MeetingEvent 会议事件文档结构
type MeetingEvent struct {
	ID        string                 `bson:"_id,omitempty" json:"id"`
	MeetingID string                 `bson:"meeting_id" json:"meeting_id"`
	UserID    string                 `bson:"user_id,omitempty" json:"user_id,omitempty"`
	EventType string                 `bson:"event_type" json:"event_type"` // join, leave, mute, unmute, share_screen, etc.
	Data      map[string]interface{} `bson:"data,omitempty" json:"data,omitempty"`
	Timestamp time.Time              `bson:"timestamp" json:"timestamp"`
}

// MongoRepository MongoDB仓库基类
type MongoRepository struct {
	collection *mongo.Collection
}

// NewMongoRepository 创建MongoDB仓库
func NewMongoRepository(collectionName string) *MongoRepository {
	if MongoDB == nil {
		panic("MongoDB not initialized")
	}
	return &MongoRepository{
		collection: MongoDB.Collection(collectionName),
	}
}

// Insert 插入文档
func (r *MongoRepository) Insert(ctx context.Context, document interface{}) (*mongo.InsertOneResult, error) {
	return r.collection.InsertOne(ctx, document)
}

// InsertMany 批量插入文档
func (r *MongoRepository) InsertMany(ctx context.Context, documents []interface{}) (*mongo.InsertManyResult, error) {
	return r.collection.InsertMany(ctx, documents)
}

// FindOne 查找单个文档
func (r *MongoRepository) FindOne(ctx context.Context, filter interface{}, result interface{}) error {
	return r.collection.FindOne(ctx, filter).Decode(result)
}

// Find 查找多个文档
func (r *MongoRepository) Find(ctx context.Context, filter interface{}, opts ...*options.FindOptions) (*mongo.Cursor, error) {
	return r.collection.Find(ctx, filter, opts...)
}

// UpdateOne 更新单个文档
func (r *MongoRepository) UpdateOne(ctx context.Context, filter interface{}, update interface{}) (*mongo.UpdateResult, error) {
	return r.collection.UpdateOne(ctx, filter, update)
}

// UpdateMany 更新多个文档
func (r *MongoRepository) UpdateMany(ctx context.Context, filter interface{}, update interface{}) (*mongo.UpdateResult, error) {
	return r.collection.UpdateMany(ctx, filter, update)
}

// DeleteOne 删除单个文档
func (r *MongoRepository) DeleteOne(ctx context.Context, filter interface{}) (*mongo.DeleteResult, error) {
	return r.collection.DeleteOne(ctx, filter)
}

// DeleteMany 删除多个文档
func (r *MongoRepository) DeleteMany(ctx context.Context, filter interface{}) (*mongo.DeleteResult, error) {
	return r.collection.DeleteMany(ctx, filter)
}

// CountDocuments 统计文档数量
func (r *MongoRepository) CountDocuments(ctx context.Context, filter interface{}) (int64, error) {
	return r.collection.CountDocuments(ctx, filter)
}

// Aggregate 聚合查询
func (r *MongoRepository) Aggregate(ctx context.Context, pipeline interface{}) (*mongo.Cursor, error) {
	return r.collection.Aggregate(ctx, pipeline)
}

// ChatMessageRepository 聊天消息仓库
type ChatMessageRepository struct {
	*MongoRepository
}

// NewChatMessageRepository 创建聊天消息仓库
func NewChatMessageRepository() *ChatMessageRepository {
	return &ChatMessageRepository{
		MongoRepository: NewMongoRepository("chat_messages"),
	}
}

// GetMessagesByMeeting 获取会议的聊天消息
func (r *ChatMessageRepository) GetMessagesByMeeting(ctx context.Context, meetingID string, limit int64, skip int64) ([]ChatMessage, error) {
	filter := map[string]interface{}{"meeting_id": meetingID}
	opts := options.Find().
		SetSort(map[string]interface{}{"timestamp": -1}).
		SetLimit(limit).
		SetSkip(skip)

	cursor, err := r.Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var messages []ChatMessage
	if err := cursor.All(ctx, &messages); err != nil {
		return nil, err
	}

	return messages, nil
}

// AIAnalysisRepository AI分析结果仓库
type AIAnalysisRepository struct {
	*MongoRepository
}

// NewAIAnalysisRepository 创建AI分析结果仓库
func NewAIAnalysisRepository() *AIAnalysisRepository {
	return &AIAnalysisRepository{
		MongoRepository: NewMongoRepository("ai_analysis_results"),
	}
}

// GetAnalysisByMeeting 获取会议的AI分析结果
func (r *AIAnalysisRepository) GetAnalysisByMeeting(ctx context.Context, meetingID string, analysisType string) ([]AIAnalysisResult, error) {
	filter := map[string]interface{}{
		"meeting_id": meetingID,
	}
	if analysisType != "" {
		filter["analysis_type"] = analysisType
	}

	opts := options.Find().SetSort(map[string]interface{}{"timestamp": -1})
	cursor, err := r.Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var results []AIAnalysisResult
	if err := cursor.All(ctx, &results); err != nil {
		return nil, err
	}

	return results, nil
}

// MeetingEventRepository 会议事件仓库
type MeetingEventRepository struct {
	*MongoRepository
}

// NewMeetingEventRepository 创建会议事件仓库
func NewMeetingEventRepository() *MeetingEventRepository {
	return &MeetingEventRepository{
		MongoRepository: NewMongoRepository("meeting_events"),
	}
}

// GetEventsByMeeting 获取会议事件
func (r *MeetingEventRepository) GetEventsByMeeting(ctx context.Context, meetingID string, eventType string) ([]MeetingEvent, error) {
	filter := map[string]interface{}{
		"meeting_id": meetingID,
	}
	if eventType != "" {
		filter["event_type"] = eventType
	}

	opts := options.Find().SetSort(map[string]interface{}{"timestamp": -1})
	cursor, err := r.Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var events []MeetingEvent
	if err := cursor.All(ctx, &events); err != nil {
		return nil, err
	}

	return events, nil
}
