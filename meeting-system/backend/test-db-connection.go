package main

import (
	"fmt"
	"log"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"meeting-system/shared/models"
)

func main() {
	fmt.Println("Testing SQLite database connection...")

	// 连接SQLite数据库
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})

	if err != nil {
		log.Fatal("Failed to connect to SQLite database:", err)
	}

	fmt.Println("✅ SQLite database connection successful")

	// 测试表迁移
	fmt.Println("Testing database migration...")
	err = db.AutoMigrate(
		&models.User{},
		&models.Meeting{},
		&models.MeetingParticipant{},
		&models.MeetingRecording{},
		&models.MediaStream{},
		&models.MeetingRoom{},
	)

	if err != nil {
		log.Fatal("Failed to migrate database:", err)
	}

	fmt.Println("✅ Database migration successful")

	// 测试创建用户
	fmt.Println("Testing user creation...")
	user := models.User{
		Username: "testuser",
		Email:    "test@example.com",
		Password: "hashedpassword",
		Nickname: "Test User",
		Status:   models.UserStatusActive,
	}

	result := db.Create(&user)
	if result.Error != nil {
		log.Fatal("Failed to create user:", result.Error)
	}

	fmt.Printf("✅ User created successfully with ID: %d\n", user.ID)

	// 测试查询用户
	var foundUser models.User
	result = db.First(&foundUser, "username = ?", "testuser")
	if result.Error != nil {
		log.Fatal("Failed to find user:", result.Error)
	}

	fmt.Printf("✅ User found: %s (%s)\n", foundUser.Username, foundUser.Email)

	fmt.Println("🎉 All database tests passed!")
}
