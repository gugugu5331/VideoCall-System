package services

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"meeting-system/media-service/models"
	"meeting-system/shared/config"
	"meeting-system/shared/storage"
)

func TestMediaServiceWithMinIOAndFFmpeg(t *testing.T) {
	// 确保 ffmpeg 可用
	if _, err := exec.LookPath("ffmpeg"); err != nil {
		t.Skip("ffmpeg not available, skipping integration test")
	}

	cfg := &config.Config{
		MinIO: config.MinIOConfig{
			Endpoint:        "127.0.0.1:9100",
			AccessKeyID:     "minioadmin",
			SecretAccessKey: "minioadmin",
			UseSSL:          false,
			BucketName:      "meeting-media",
		},
	}

	if err := storage.InitMinIO(cfg.MinIO); err != nil {
		t.Skipf("minio not available: %v", err)
	}

	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	require.NoError(t, err)

	require.NoError(t, db.AutoMigrate(
		&models.MediaFile{},
		&models.ProcessingJob{},
		&models.Recording{},
		// SFU 架构：Stream 模型已移除（用于RTMP/HLS）
		&models.WebRTCPeer{},
		// SFU 架构：Filter 模型已移除
		&models.MediaStats{},
	))

	mediaService := NewMediaService(cfg, db, nil)
	mediaService.SetStorageClient(NewMinIOStorageAdapter(cfg.MinIO.BucketName, 10*time.Second))
	require.NoError(t, mediaService.Initialize())

	ffmpegService := NewFFmpegService(cfg, mediaService, nil)
	require.NoError(t, ffmpegService.Initialize())
	defer ffmpegService.Stop()

	// 生成测试视频
	sampleFile := filepath.Join(os.TempDir(), fmt.Sprintf("sample-%d.mp4", time.Now().UnixNano()))
	cmd := exec.Command("ffmpeg", "-y",
		"-f", "lavfi", "-i", "color=c=blue:s=320x240:d=1",
		"-f", "lavfi", "-i", "sine=frequency=1000:duration=1",
		"-shortest",
		"-movflags", "+faststart",
		"-c:v", "libx264",
		"-pix_fmt", "yuv420p",
		"-c:a", "aac",
		sampleFile,
	)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	require.NoError(t, cmd.Run())
	defer os.Remove(sampleFile)

	file, err := os.Open(sampleFile)
	require.NoError(t, err)
	defer file.Close()

	stat, err := file.Stat()
	require.NoError(t, err)

	storagePath := fmt.Sprintf("media/test/%s.mp4", uuid.NewString())
	adapter := NewMinIOStorageAdapter(cfg.MinIO.BucketName, 10*time.Second)
	require.NoError(t, adapter.UploadFile("media", storagePath, file, stat.Size(), "video/mp4"))

	mediaFile := &models.MediaFile{
		FileID:       uuid.NewString(),
		FileName:     filepath.Base(storagePath),
		OriginalName: "sample.mp4",
		FileType:     "video",
		MimeType:     "video/mp4",
		FileSize:     stat.Size(),
		StoragePath:  storagePath,
		Status:       "uploaded",
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}
	require.NoError(t, db.Create(mediaFile).Error)

	jobID, err := ffmpegService.GenerateThumbnail(mediaFile.FileID, 0.5)
	require.NoError(t, err)

	var job *ProcessingJob
	deadline := time.Now().Add(30 * time.Second)
	for {
		job, err = ffmpegService.GetJobStatus(jobID)
		require.NoError(t, err)
		if job.Status == "completed" {
			break
		}
		if job.Status == "failed" {
			t.Fatalf("ffmpeg job failed: %s", job.Error)
		}
		if time.Now().After(deadline) {
			t.Fatalf("ffmpeg job did not finish in time; status=%s", job.Status)
		}
		time.Sleep(500 * time.Millisecond)
	}

	require.Equal(t, "completed", job.Status)

	// 验证缩略图已上传
	thumbObject := fmt.Sprintf("processed/%s", filepath.Base(job.OutputPath))
	reader, err := adapter.GetFile("media", thumbObject)
	require.NoError(t, err)
	defer reader.Close()

	buf := make([]byte, 1)
	_, err = reader.Read(buf)
	require.NoError(t, err)

	// 清理对象
	require.NoError(t, adapter.DeleteFile("media", storagePath))
	require.NoError(t, adapter.DeleteFile("media", thumbObject))
}
