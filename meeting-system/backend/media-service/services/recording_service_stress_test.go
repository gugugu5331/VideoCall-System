package services

import (
	"fmt"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"meeting-system/media-service/models"
	"meeting-system/shared/config"
)

func TestRecordingServiceStress(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping stress test in short mode")
	}

	// 初始化内存数据库
	// SQLite 并发优化配置：
	// - mode=memory：使用内存数据库（每次测试独立实例）
	// - _busy_timeout=5000：等待锁释放的超时时间（5秒）
	// - _journal_mode=WAL：Write-Ahead Logging 模式，允许并发读写
	// 注意：不使用 cache=shared 以确保每次测试运行使用独立的数据库实例
	db, err := gorm.Open(sqlite.Open("file::memory:?mode=memory&_busy_timeout=5000&_journal_mode=WAL"), &gorm.Config{})
	require.NoError(t, err)

	// 配置连接池以优化并发性能
	// MaxOpenConns=1：强制串行化写入，避免 "database is locked" 错误
	// MaxIdleConns=1：保持一个空闲连接
	// ConnMaxLifetime=0：连接不过期
	sqlDB, err := db.DB()
	require.NoError(t, err)
	sqlDB.SetMaxOpenConns(1)
	sqlDB.SetMaxIdleConns(1)
	sqlDB.SetConnMaxLifetime(0)

	err = db.AutoMigrate(
		&models.MediaFile{},
		&models.ProcessingJob{},
		&models.Recording{},
		// SFU 架构：Stream 模型已移除（用于RTMP/HLS）
		&models.WebRTCPeer{},
		// SFU 架构：Filter 模型已移除
		&models.MediaStats{},
	)
	require.NoError(t, err)

	cfg := &config.Config{
		Signaling: config.SignalingConfig{
			Session: config.SessionConfig{ConnectionTimeout: 60},
		},
	}

	mediaService := NewMediaService(cfg, db, nil)
	require.NoError(t, mediaService.Initialize())

	recordingService := NewRecordingService(cfg, mediaService, nil, nil)
	require.NoError(t, recordingService.Initialize())
	defer recordingService.Stop()

	var (
		concurrency      = 20
		iterationsPerGor = 3
		successfulStarts atomic.Int64
		startFailures    atomic.Int64
		stopFailures     atomic.Int64
		wg               sync.WaitGroup
	)

	startTime := time.Now()

	for i := 0; i < concurrency; i++ {
		wg.Add(1)
		go func(worker int) {
			defer wg.Done()
			for j := 0; j < iterationsPerGor; j++ {
				req := &StartRecordingRequest{
					MeetingID:    fmt.Sprintf("meeting-%d", worker),
					RoomID:       fmt.Sprintf("room-%d", worker),
					UserID:       fmt.Sprintf("user-%d", worker),
					Title:        fmt.Sprintf("Stress Recording %d-%d", worker, j),
					Quality:      "720p",
					Format:       "mp4",
					Participants: []string{fmt.Sprintf("user-%d", worker)},
				}

				recording, err := recordingService.StartRecording(req)
				if err != nil {
					startFailures.Add(1)
					continue
				}

				successfulStarts.Add(1)

				// 模拟在房间短暂存在
				time.Sleep(20 * time.Millisecond)

				if err := recordingService.StopRecording(recording.RecordingID); err != nil {
					stopFailures.Add(1)
				}
			}
		}(i)
	}

	wg.Wait()
	elapsed := time.Since(startTime)

	// 等待后台录制流程完成
	time.Sleep(6 * time.Second)

	// 验证数据库中的录制记录
	var total int64
	require.NoError(t, db.Model(&models.Recording{}).Count(&total).Error)
	require.Equal(t, successfulStarts.Load(), total, "unexpected number of recordings persisted")

	var incomplete []models.Recording
	require.NoError(t, db.Where("status <> ?", "completed").Find(&incomplete).Error)
	require.Empty(t, incomplete, "all recordings should be marked completed after cleanup")

	t.Logf("media stress: %d started, %d start failures, %d stop failures, duration=%s", successfulStarts.Load(), startFailures.Load(), stopFailures.Load(), elapsed)
}
