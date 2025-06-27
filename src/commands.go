package main

import (
	"context"
	"fmt"
	"strings"

	"go.mau.fi/whatsmeow"
	"go.mau.fi/whatsmeow/proto/waE2E"
	"go.mau.fi/whatsmeow/types/events"
	"google.golang.org/protobuf/proto"
)

func handlePing(client *whatsmeow.Client, msg *events.Message, jobs chan<- DownloadJob) {
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

func handleDownload(client *whatsmeow.Client, msg *events.Message, jobs chan<- DownloadJob) {
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
	resp, _ := client.SendMessage(context.Background(), msg.Info.Chat, &waE2E.Message{
		Conversation: proto.String(waitMsg),
	})

	newJob := DownloadJob{
		UserInfo:     msg.Info,
		VideoURL:     parts[1],
		InitialMsgID: resp.ID,
	}

	jobs <- newJob

	fmt.Printf("[HANDLER] Job for %s queued\n", msg.Info.Chat.User)
}
