package database

import (
	"context"
	"fmt"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// MongoDB MongoDB客户端
type MongoDB struct {
	Client   *mongo.Client
	Database *mongo.Database
}

// NewMongoDB 创建MongoDB连接
func NewMongoDB(uri, dbName string) (*MongoDB, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(uri))
	if err != nil {
		return nil, fmt.Errorf("failed to connect to MongoDB: %w", err)
	}

	// 测试连接
	if err := client.Ping(ctx, nil); err != nil {
		return nil, fmt.Errorf("failed to ping MongoDB: %w", err)
	}

	database := client.Database(dbName)
	log.Println("Successfully connected to MongoDB")

	return &MongoDB{
		Client:   client,
		Database: database,
	}, nil
}

// Close 关闭MongoDB连接
func (m *MongoDB) Close() error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	return m.Client.Disconnect(ctx)
}

// Collection 获取集合
func (m *MongoDB) Collection(name string) *mongo.Collection {
	return m.Database.Collection(name)
}
