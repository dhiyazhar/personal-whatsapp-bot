package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"image"
	"image/jpeg"
	_ "image/png"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/nfnt/resize"
	_ "golang.org/x/image/webp"
)

type VideoInfo struct {
	ID          string `json:"id"`
	Title       string `json:"title"`
	Duration    int    `json:"duration"`
	Uploader    string `json:"uploader"`
	ViewCount   int64  `json:"view_count"`
	Thumbnail   string `json:"thumbnail"`
	Description string `json:"description"`
	Filename    string `json:"_filename"`
}

func youtubeVideoDownload(ctx context.Context, videoURL string) (string, error) {
	tmpDir := os.TempDir()

	maxRetries := 3
	var lastError error

	for attempt := 1; attempt <= maxRetries; attempt++ {
		tmpFile := filepath.Join(tmpDir, fmt.Sprintf("video_%d_%d.mp4", time.Now().UnixNano(), attempt))

		ytdlp := exec.CommandContext(ctx, "yt-dlp",
			"-f", "bestvideo[height<=720]+bestaudio/best",
			"-S", "vcodec:h264,res,acodec:m4a",
			"--no-playlist",
			"-o", tmpFile,
			videoURL,
		)

		out, err := ytdlp.CombinedOutput()
		if err != nil {
			lastError = fmt.Errorf("yt-dlp error: %v, %s", err, out)

			if _, statErr := os.Stat(tmpFile); statErr == nil {
				os.Remove(tmpFile)
			}

			if attempt < maxRetries {
				waitTime := time.Duration(attempt) * 2 * time.Second

				select {
				case <-ctx.Done():
					return "", ctx.Err()
				case <-time.After(waitTime):

				}
				continue
			}
		} else {
			if fileInfo, statErr := os.Stat(tmpFile); statErr == nil && fileInfo.Size() > 0 {
				fmt.Printf("[DEBUG] download successful on attempt %d\n", attempt)
				return tmpFile, nil
			} else {
				lastError = fmt.Errorf("download complete but file is missing or empty on attempt %d", attempt)

				os.Remove(tmpFile)

				if attempt < maxRetries {

					waitTime := time.Duration(attempt) * 2 * time.Second
					select {
					case <-ctx.Done():
						return "", ctx.Err()
					case <-time.After(waitTime):
						continue
					}

				}
			}
		}

	}

	return "", fmt.Errorf("all %d attempts failed, last error: %v", maxRetries, lastError)
}

func getVideoInfo(videoURL string) (*VideoInfo, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, "yt-dlp",
		"--dump-json",
		"--no-download",
		videoURL)

	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to get video info: %v", err)
	}

	var info VideoInfo
	if err := json.Unmarshal(output, &info); err != nil {
		return nil, fmt.Errorf("failed to parse video info: %v", err)
	}

	return &info, nil
}

func getVideoThumbnail(thumbnailURL string) ([]byte, error) {
	resp, err := http.Get(thumbnailURL)
	if err != nil {
		return nil, fmt.Errorf("failed to download thumbnail: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("bad status code for thumbnail download: %s", resp.Status)
	}

	img, format, err := image.Decode(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to decode thumbnail image: %w", err)
	}
	fmt.Printf("[DEBUG] decoded thumbnail format: %s\n", format)

	resizedImg := resize.Resize(72, 0, img, resize.Lanczos3)

	var buf bytes.Buffer
	err = jpeg.Encode(&buf, resizedImg, &jpeg.Options{Quality: 90})
	if err != nil {
		return nil, fmt.Errorf("failed to encode thumbnail image: %w", err)
	}

	return buf.Bytes(), nil
}
