package storage

import (
	"context"
	"fmt"
	"io"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"

	"meeting-system/shared/config"
	"meeting-system/shared/logger"
)

var MinIOClient *minio.Client

// InitMinIO 初始化MinIO客户端
func InitMinIO(config config.MinIOConfig) error {
	client, err := minio.New(config.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(config.AccessKeyID, config.SecretAccessKey, ""),
		Secure: config.UseSSL,
	})
	if err != nil {
		return fmt.Errorf("failed to create MinIO client: %w", err)
	}

	MinIOClient = client

	// 检查连接
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// 检查存储桶是否存在，不存在则创建
	exists, err := client.BucketExists(ctx, config.BucketName)
	if err != nil {
		return fmt.Errorf("failed to check bucket existence: %w", err)
	}

	if !exists {
		err = client.MakeBucket(ctx, config.BucketName, minio.MakeBucketOptions{})
		if err != nil {
			return fmt.Errorf("failed to create bucket: %w", err)
		}
		logger.Info("MinIO bucket created", logger.String("bucket", config.BucketName))
	}

	logger.Info("MinIO connected successfully",
		logger.String("endpoint", config.Endpoint),
		logger.String("bucket", config.BucketName))

	return nil
}

// GetMinIOClient 获取MinIO客户端
func GetMinIOClient() *minio.Client {
	return MinIOClient
}

// StorageService 存储服务
type StorageService struct {
	client     *minio.Client
	bucketName string
}

// NewStorageService 创建存储服务
func NewStorageService(bucketName string) *StorageService {
	return &StorageService{
		client:     MinIOClient,
		bucketName: bucketName,
	}
}

// UploadFile 上传文件
func (s *StorageService) UploadFile(ctx context.Context, objectName string, reader io.Reader, objectSize int64, contentType string) (*minio.UploadInfo, error) {
	uploadInfo, err := s.client.PutObject(ctx, s.bucketName, objectName, reader, objectSize, minio.PutObjectOptions{
		ContentType: contentType,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to upload file: %w", err)
	}

	logger.Info("File uploaded successfully",
		logger.String("bucket", s.bucketName),
		logger.String("object", objectName),
		logger.Int64("size", objectSize))

	return &uploadInfo, nil
}

// DownloadFile 下载文件
func (s *StorageService) DownloadFile(ctx context.Context, objectName string) (*minio.Object, error) {
	object, err := s.client.GetObject(ctx, s.bucketName, objectName, minio.GetObjectOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to download file: %w", err)
	}

	return object, nil
}

// DeleteFile 删除文件
func (s *StorageService) DeleteFile(ctx context.Context, objectName string) error {
	err := s.client.RemoveObject(ctx, s.bucketName, objectName, minio.RemoveObjectOptions{})
	if err != nil {
		return fmt.Errorf("failed to delete file: %w", err)
	}

	logger.Info("File deleted successfully",
		logger.String("bucket", s.bucketName),
		logger.String("object", objectName))

	return nil
}

// GetFileInfo 获取文件信息
func (s *StorageService) GetFileInfo(ctx context.Context, objectName string) (*minio.ObjectInfo, error) {
	objectInfo, err := s.client.StatObject(ctx, s.bucketName, objectName, minio.StatObjectOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to get file info: %w", err)
	}

	return &objectInfo, nil
}

// ListFiles 列出文件
func (s *StorageService) ListFiles(ctx context.Context, prefix string, recursive bool) ([]minio.ObjectInfo, error) {
	var objects []minio.ObjectInfo

	objectCh := s.client.ListObjects(ctx, s.bucketName, minio.ListObjectsOptions{
		Prefix:    prefix,
		Recursive: recursive,
	})

	for object := range objectCh {
		if object.Err != nil {
			return nil, fmt.Errorf("failed to list files: %w", object.Err)
		}
		objects = append(objects, object)
	}

	return objects, nil
}

// GeneratePresignedURL 生成预签名URL
func (s *StorageService) GeneratePresignedURL(ctx context.Context, objectName string, expiry time.Duration) (string, error) {
	url, err := s.client.PresignedGetObject(ctx, s.bucketName, objectName, expiry, nil)
	if err != nil {
		return "", fmt.Errorf("failed to generate presigned URL: %w", err)
	}

	return url.String(), nil
}

// CopyFile 复制文件
func (s *StorageService) CopyFile(ctx context.Context, srcObjectName, destObjectName string) error {
	srcOpts := minio.CopySrcOptions{
		Bucket: s.bucketName,
		Object: srcObjectName,
	}

	destOpts := minio.CopyDestOptions{
		Bucket: s.bucketName,
		Object: destObjectName,
	}

	_, err := s.client.CopyObject(ctx, destOpts, srcOpts)
	if err != nil {
		return fmt.Errorf("failed to copy file: %w", err)
	}

	logger.Info("File copied successfully",
		logger.String("bucket", s.bucketName),
		logger.String("src", srcObjectName),
		logger.String("dest", destObjectName))

	return nil
}

// FileUploadResult 文件上传结果
type FileUploadResult struct {
	ObjectName  string `json:"object_name"`
	Size        int64  `json:"size"`
	ContentType string `json:"content_type"`
	URL         string `json:"url"`
	ETag        string `json:"etag"`
}

// UploadAvatar 上传用户头像
func (s *StorageService) UploadAvatar(ctx context.Context, userID uint, reader io.Reader, size int64, contentType string) (*FileUploadResult, error) {
	objectName := fmt.Sprintf("users/%d/avatars/%d.jpg", userID, time.Now().Unix())

	uploadInfo, err := s.UploadFile(ctx, objectName, reader, size, contentType)
	if err != nil {
		return nil, err
	}

	// 生成访问URL（24小时有效）
	url, err := s.GeneratePresignedURL(ctx, objectName, 24*time.Hour)
	if err != nil {
		logger.Warn("Failed to generate presigned URL", logger.Err(err))
		url = fmt.Sprintf("/api/v1/files/%s", objectName)
	}

	return &FileUploadResult{
		ObjectName:  objectName,
		Size:        size,
		ContentType: contentType,
		URL:         url,
		ETag:        uploadInfo.ETag,
	}, nil
}

// UploadMeetingFile 上传会议文件
func (s *StorageService) UploadMeetingFile(ctx context.Context, meetingID uint, filename string, reader io.Reader, size int64, contentType string) (*FileUploadResult, error) {
	objectName := fmt.Sprintf("meetings/%d/files/%d_%s", meetingID, time.Now().Unix(), filename)

	uploadInfo, err := s.UploadFile(ctx, objectName, reader, size, contentType)
	if err != nil {
		return nil, err
	}

	// 生成访问URL（7天有效）
	url, err := s.GeneratePresignedURL(ctx, objectName, 7*24*time.Hour)
	if err != nil {
		logger.Warn("Failed to generate presigned URL", logger.Err(err))
		url = fmt.Sprintf("/api/v1/files/%s", objectName)
	}

	return &FileUploadResult{
		ObjectName:  objectName,
		Size:        size,
		ContentType: contentType,
		URL:         url,
		ETag:        uploadInfo.ETag,
	}, nil
}

// UploadRecording 上传会议录制
func (s *StorageService) UploadRecording(ctx context.Context, meetingID uint, filename string, reader io.Reader, size int64) (*FileUploadResult, error) {
	objectName := fmt.Sprintf("meetings/%d/recordings/%s", meetingID, filename)

	uploadInfo, err := s.UploadFile(ctx, objectName, reader, size, "video/mp4")
	if err != nil {
		return nil, err
	}

	return &FileUploadResult{
		ObjectName:  objectName,
		Size:        size,
		ContentType: "video/mp4",
		URL:         fmt.Sprintf("/api/v1/recordings/%s", objectName),
		ETag:        uploadInfo.ETag,
	}, nil
}

// DeleteMeetingFiles 删除会议相关文件
func (s *StorageService) DeleteMeetingFiles(ctx context.Context, meetingID uint) error {
	prefix := fmt.Sprintf("meetings/%d/", meetingID)

	objects, err := s.ListFiles(ctx, prefix, true)
	if err != nil {
		return err
	}

	for _, object := range objects {
		if err := s.DeleteFile(ctx, object.Key); err != nil {
			logger.Warn("Failed to delete file",
				logger.String("object", object.Key),
				logger.Err(err))
		}
	}

	logger.Info("Meeting files deleted",
		logger.Uint("meeting_id", meetingID),
		logger.Int("file_count", len(objects)))

	return nil
}

// GetMeetingFiles 获取会议文件列表
func (s *StorageService) GetMeetingFiles(ctx context.Context, meetingID uint) ([]minio.ObjectInfo, error) {
	prefix := fmt.Sprintf("meetings/%d/files/", meetingID)
	return s.ListFiles(ctx, prefix, false)
}

// GetMeetingRecordings 获取会议录制列表
func (s *StorageService) GetMeetingRecordings(ctx context.Context, meetingID uint) ([]minio.ObjectInfo, error) {
	prefix := fmt.Sprintf("meetings/%d/recordings/", meetingID)
	return s.ListFiles(ctx, prefix, false)
}
