package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"go.mau.fi/whatsmeow"
	"go.mau.fi/whatsmeow/proto/waE2E"
	"google.golang.org/protobuf/proto"
)

func StartDownloadWorker(client *whatsmeow.Client, jobs <-chan DownloadJob) {
	fmt.Println("[WORKER] Download worker started")
	for job := range jobs {
		processJob(client, job) // Extract ke function terpisah
	}
}

func processJob(client *whatsmeow.Client, job DownloadJob) {
	fmt.Printf("[WORKER] starting job for user %s , URL: %s\n", job.UserInfo.Chat.User, job.VideoURL)
	start := time.Now()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel() // ✅ Best practice: selalu defer cancel setelah WithTimeout

	info, err := getVideoInfo(job.VideoURL)
	if err != nil {
		fmt.Printf("[WORKER ERR] error getting video info: %v\n", err)
		sendWorkerMessage(client, job.UserInfo.Chat, "error saat proses download")
		return
	}

	if info.Duration > 300 {
		fmt.Printf("[WORKER INFO] Video duration exceeded")
		sendWorkerMessage(client, job.UserInfo.Chat, "Durasi video melebihi limit pesan WhatsApp!")
		return
	}

	filePath, err := youtubeVideoDownload(ctx, job.VideoURL)
	if err != nil {
		fmt.Printf("[WORKER ERR] Gagal download video: %v\n", err)
		sendWorkerMessage(client, job.UserInfo.Chat, "error saat proses download")
		return
	}
	defer os.Remove(filePath) // ✅ Sekarang akan cleanup di akhir function

	// Thumbnail handling
	var thumbnailData []byte
	if info.Thumbnail != "" {
		fmt.Println("[DEBUG] creating thumbnail from URL: ", info.Thumbnail)
		thumbnailData, err = getVideoThumbnail(info.Thumbnail)
		if err != nil {
			fmt.Printf("[DEBUG] failed to create thumbnail from URL: %v\n", err)
		}
	} else {
		fmt.Println("[DEBUG] No thumbnail URL found in the video")
	}

	data, err := os.ReadFile(filePath)
	if err != nil {
		sendWorkerMessage(client, job.UserInfo.Chat, "error saat proses download")
		return
	}

	uploaded, err := client.Upload(context.Background(), data, whatsmeow.MediaVideo)
	if err != nil {
		sendWorkerMessage(client, job.UserInfo.Chat, "error saat proses download")
		return
	}

	videoMessage := &waE2E.Message{
		VideoMessage: &waE2E.VideoMessage{
			URL:           proto.String(uploaded.URL),
			DirectPath:    proto.String(uploaded.DirectPath),
			Mimetype:      proto.String("video/mp4"),
			FileEncSHA256: uploaded.FileEncSHA256,
			FileSHA256:    uploaded.FileSHA256,
			FileLength:    proto.Uint64(uploaded.FileLength),
			MediaKey:      uploaded.MediaKey,
			Caption:       proto.String(fmt.Sprintf("%s\n\n%s", info.Title, job.VideoURL)),
			Seconds:       proto.Uint32(uint32(info.Duration)),
			JPEGThumbnail: thumbnailData,
		},
	}

	_, err = client.SendMessage(context.Background(), job.UserInfo.Chat, videoMessage)
	if err != nil {
		fmt.Printf("[ERR] gagal mengirim pesan video: %s\n", err)
	} else {
		fmt.Printf("[INFO] sukses mengirim pesan video: %s\n", info.Title)
	}

	dur := time.Since(start)
	fmt.Printf("[WORKER] finished job for %s in %s\n", job.UserInfo.Chat.User, dur)
}
