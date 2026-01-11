# Test Video/Audio Assets

This directory is reserved for local test media used by the media service (upload/download, recording, FFmpeg helpers). Files are **not** tracked in Git.

## Usage

- Place small demo files here when running integration tests locally.
- Suggested formats: MP4/MOV/MKV for video; MP3/WAV/AAC for audio.
- Keep files under ~50MB and 10–60 seconds to speed up pipelines.

## Generate samples with FFmpeg

```bash
# 10s color bars + tone
ffmpeg -f lavfi -i testsrc=duration=10:size=1280x720:rate=30 \
       -f lavfi -i sine=frequency=1000:duration=10 \
       -pix_fmt yuv420p test_video.mp4

# 10s sine wave
ffmpeg -f lavfi -i sine=frequency=1000:duration=10 test_audio.mp3
```

For CI/CD, download test files from external storage or generate them on the fly to avoid repository bloat.

Cleanup tip: remove large artifacts after local runs to keep the workspace small (files are gitignored).
