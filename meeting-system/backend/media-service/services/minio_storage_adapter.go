package services

import (
	"context"
	"fmt"
	"io"
	"time"

	"meeting-system/shared/logger"
	"meeting-system/shared/storage"
)

// MinIOStorageAdapter 将共享存储服务封装成 MediaService 可使用的 StorageClient
// 默认使用单一 bucket，通过对象路径前缀区分业务目录。
type MinIOStorageAdapter struct {
	service *storage.StorageService
	bucket  string
	timeout time.Duration
}

// NewMinIOStorageAdapter 创建 MinIO 适配器，timeout<=0 时默认 30s。
func NewMinIOStorageAdapter(bucket string, timeout time.Duration) *MinIOStorageAdapter {
	svc := storage.NewStorageService(bucket)
	if timeout <= 0 {
		timeout = 30 * time.Second
	}
	return &MinIOStorageAdapter{
		service: svc,
		bucket:  bucket,
		timeout: timeout,
	}
}

func (a *MinIOStorageAdapter) contextWithTimeout() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), a.timeout)
}

func (a *MinIOStorageAdapter) ensureBucket(requestBucket string) {
	if requestBucket != "" && requestBucket != a.bucket {
		logger.Warn("Ignoring non-default bucket for MinIO adapter",
			logger.String("requested_bucket", requestBucket),
			logger.String("adapter_bucket", a.bucket),
		)
	}
}

// UploadFile 上传对象到 MinIO。
func (a *MinIOStorageAdapter) UploadFile(bucket, object string, reader io.Reader, size int64, contentType string) error {
	a.ensureBucket(bucket)
	ctx, cancel := a.contextWithTimeout()
	defer cancel()

	if _, err := a.service.UploadFile(ctx, object, reader, size, contentType); err != nil {
		return fmt.Errorf("minio upload failed: %w", err)
	}
	return nil
}

// GetFile 从 MinIO 读取对象。
func (a *MinIOStorageAdapter) GetFile(bucket, object string) (io.ReadCloser, error) {
	a.ensureBucket(bucket)
	ctx, cancel := a.contextWithTimeout()
	file, err := a.service.DownloadFile(ctx, object)
	if err != nil {
		cancel()
		return nil, fmt.Errorf("minio download failed: %w", err)
	}
	return &minioObjectReader{ReadCloser: file, cancel: cancel}, nil
}

// DeleteFile 删除 MinIO 对象。
func (a *MinIOStorageAdapter) DeleteFile(bucket, object string) error {
	a.ensureBucket(bucket)
	ctx, cancel := a.contextWithTimeout()
	defer cancel()

	if err := a.service.DeleteFile(ctx, object); err != nil {
		return fmt.Errorf("minio delete failed: %w", err)
	}
	return nil
}

type minioObjectReader struct {
	io.ReadCloser
	cancel context.CancelFunc
}

func (r *minioObjectReader) Close() error {
	if r.cancel != nil {
		r.cancel()
	}
	return r.ReadCloser.Close()
}
