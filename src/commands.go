package main

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"go.mau.fi/whatsmeow"
	"go.mau.fi/whatsmeow/proto/waE2E"
	"go.mau.fi/whatsmeow/types/events"
	"google.golang.org/protobuf/proto"
)

func handlePing(client *whatsmeow.Client, msg *events.Message) {
	text := extractTextFromMsg(msg)
	if text == "" {
		fmt.Println("[DEBUG] no text in ping handler")
		return
	}

	if text != "!ping" {
		return
	}

	sendRespondMsgAsync(client, msg, "Pong!")

}

func handleDownload(client *whatsmeow.Client, msg *events.Message) {
	text := extractTextFromMsg(msg)
	if text == "" {
		fmt.Println("[DEBUG] no text in download handler")
		return
	}

	parts := strings.Fields(text)
	if len(parts) > 2 {
		fmt.Println("[DEBUG] terlalu banyak parameter")

		sendRespondMsgAsync(client, msg, "terlalu banyak parameter")

		return
	} else if len(parts) < 2 {
		if strings.ToLower(parts[0]) != "!download" {
			fmt.Println("[DEBUG] command bukan download!")
			return
		}

		fmt.Println("[DEBUG] command butuh parameter link")

		sendRespondMsgAsync(client, msg, "command butuh parameter link")

		return
	}

	waitMsg := "Mohon tunggu sebentar...\n\nBot hanya support 720p untuk sementara"
	sendRespondMsgAsync(client, msg, waitMsg)

	go func() {
		start := time.Now()
		defer func() {
			dur := time.Since(start)
			fmt.Printf("[METRICS] goroutine total execution time: %s\n", dur)
		}()

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
		defer cancel()

		videoURL := parts[1]
		info, err := getVideoInfo(videoURL)
		if err != nil {
			fmt.Printf("[ERR] error getting video info: %v\n", err)
			return
		}

		if info.Duration > 300 {
			fmt.Printf("[DEBUG] video info duration: %v\n", info.Duration)
			sendRespondMsgAsync(client, msg, "durasi video melebihi limit WhatsApp")
			return
		}

		filePath, err := youtubeVideoDownload(ctx, videoURL)
		if err != nil {
			sendRespondMsgAsync(client, msg, fmt.Sprintf("Error download: %v", err))
			return
		}
		defer os.Remove(filePath)

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
			sendRespondMsgAsync(client, msg, "Gagal baca file video")
			return
		}

		uploaded, err := client.Upload(context.Background(), data, whatsmeow.MediaVideo)
		if err != nil {
			sendRespondMsgAsync(client, msg, "error saat proses download")
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
				Caption:       proto.String(fmt.Sprintf("%s\n\n%s", info.Title, videoURL)),
				Seconds:       proto.Uint32(uint32(info.Duration)),
				JPEGThumbnail: thumbnailData,
			},
		}

		_, err = client.SendMessage(context.Background(), msg.Info.Chat, videoMessage)
		if err != nil {
			fmt.Printf("[ERR] gagal mengirim pesan video: %s\n", err)
		} else {
			fmt.Printf("[INFO] sukses mengirim pesan video: %s\n", info.Title)
		}
	}()
}
