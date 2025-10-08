# Test Video Files

This directory contains test video and audio files for media service testing.

## Purpose

These files are used for:
- Media upload/download testing
- Video processing testing
- FFmpeg integration testing
- Recording functionality testing

## File Requirements

Test files should be:
- **Video formats**: MP4, AVI, MOV, MKV
- **Audio formats**: MP3, WAV, AAC
- **Size**: Keep test files under 50MB each
- **Duration**: 10-60 seconds recommended

## Note

Test media files are **not included in the Git repository** due to their large size.

### To add test files:

1. Place your test video/audio files in this directory
2. Files will be automatically ignored by Git (see `.gitignore`)
3. For CI/CD, download test files from external storage or generate them programmatically

### Sample test file generation:

```bash
# Generate a test video using FFmpeg
ffmpeg -f lavfi -i testsrc=duration=10:size=1280x720:rate=30 \
       -f lavfi -i sine=frequency=1000:duration=10 \
       -pix_fmt yuv420p test_video.mp4

# Generate a test audio file
ffmpeg -f lavfi -i sine=frequency=1000:duration=10 test_audio.mp3
```

## Existing Test Files (Not in Git)

The following files may exist locally but are not tracked:
- `20250602_215504.mp3` - Audio test file
- `20250827_093242.mp4` - Video test file
- `20250827_104938.mp4` - Video test file
- `20250827_105955.mp4` - Video test file
- `20250928_164722.mp4` - Video test file
- `20250928_164800.mp4` - Video test file
- `20250928_165500.mp4` - Video test file

## Alternative: Use External Test Files

For production testing, consider using:
- Public domain test videos from [Sample Videos](https://sample-videos.com/)
- Generate test files on-the-fly in test scripts
- Store test files in cloud storage (S3, MinIO, etc.)

