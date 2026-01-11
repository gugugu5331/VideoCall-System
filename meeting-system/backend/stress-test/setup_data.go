//go:build tools
// +build tools

package main

import (
	"flag"
	"fmt"
	"log"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"meeting-system/shared/config"
	"meeting-system/shared/database"
	"meeting-system/shared/models"
)

func main() {
	var (
		configPath string
		meetingID  uint
		userCount  int
		creatorID  uint
		title      string
	)

	flag.StringVar(&configPath, "config", "config/signaling-service.yaml", "path to signaling service config")
	flag.UintVar(&meetingID, "meeting", 9999, "meeting ID for stress test")
	flag.IntVar(&userCount, "users", 200, "number of stress test users to seed")
	flag.UintVar(&creatorID, "creator", 1, "user ID to set as meeting creator")
	flag.StringVar(&title, "title", "Stress Test Meeting", "title for the seeded meeting")
	flag.Parse()

	log.Printf("loading config from %s", configPath)
	config.InitConfig(configPath)
	cfg := config.GlobalConfig

	if err := database.InitDB(cfg.Database); err != nil {
		log.Fatalf("failed to init database: %v", err)
	}
	defer database.CloseDB()

	db := database.GetDB()

	seedUsers(db, userCount)
	seedMeeting(db, meetingID, creatorID, title)
	seedParticipants(db, meetingID, userCount)

	log.Printf("seed completed: meeting %d with %d participants", meetingID, userCount)
}

func seedUsers(db *gorm.DB, userCount int) {
	log.Printf("ensuring %d users exist", userCount)
	for i := 1; i <= userCount; i++ {
		userID := uint(i)
		user := models.User{
			ID:       userID,
			Username: fmt.Sprintf("stress_user_%d", userID),
			Email:    fmt.Sprintf("stress%d@example.com", userID),
			Password: "stress-password", // test-only placeholder
			Status:   models.UserStatusActive,
		}
		if err := db.Clauses(clause.OnConflict{DoNothing: true}).Create(&user).Error; err != nil {
			log.Fatalf("failed to upsert user %d: %v", userID, err)
		}
	}
}

func seedMeeting(db *gorm.DB, meetingID, creatorID uint, title string) {
	log.Printf("ensuring meeting %d exists", meetingID)
	meeting := models.Meeting{
		ID:          meetingID,
		Title:       title,
		Description: "Auto-generated for signaling stress test",
		CreatorID:   creatorID,
		StartTime:   time.Now(),
		EndTime:     time.Now().Add(2 * time.Hour),
		Status:      models.MeetingStatusScheduled,
		MeetingType: models.MeetingTypeVideo,
		Settings:    "{}",
	}
	if err := db.Clauses(clause.OnConflict{DoNothing: true}).Create(&meeting).Error; err != nil {
		log.Fatalf("failed to upsert meeting %d: %v", meetingID, err)
	}
}

func seedParticipants(db *gorm.DB, meetingID uint, userCount int) {
	log.Printf("ensuring participants for meeting %d", meetingID)
	for i := 1; i <= userCount; i++ {
		participant := models.MeetingParticipant{
			MeetingID: meetingID,
			UserID:    uint(i),
			Role:      models.ParticipantRoleParticipant,
			Status:    models.ParticipantStatusJoined,
		}
		if err := db.Clauses(clause.OnConflict{DoNothing: true}).Create(&participant).Error; err != nil {
			log.Fatalf("failed to upsert participant meeting=%d user=%d: %v", meetingID, i, err)
		}
	}
}
