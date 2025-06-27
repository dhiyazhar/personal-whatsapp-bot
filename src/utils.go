package main

import (
	"context"
	"fmt"

	"go.mau.fi/whatsmeow"
	"go.mau.fi/whatsmeow/proto/waE2E"
	"go.mau.fi/whatsmeow/types/events"
	"google.golang.org/protobuf/proto"
)

func extractTextFromMsg(msg *events.Message) string {
	if msg == nil || msg.Message == nil {
		return ""
	}

	if conversation := msg.Message.GetConversation(); conversation != "" {
		return conversation
	}

	if msg.Message.ExtendedTextMessage != nil {
		if extText := msg.Message.ExtendedTextMessage.GetText(); extText != "" {
			return extText
		}
	}

	return ""
}

func sendRespondMsgAsync(client *whatsmeow.Client, msg *events.Message, text string) {
	go func() {

		reply := &waE2E.Message{
			Conversation: proto.String(text),
		}

		resp, err := client.SendMessage(context.Background(), msg.Info.Chat, reply)
		if err != nil {
			fmt.Printf("[WARN] gagal mengirim respon: %v\n", err)
			return
		}
		fmt.Printf("[INFO] pesan terkirim ke %s: ID=%s, Timestamp=%v\n", msg.Info.Chat.String(), resp.ID, resp.Timestamp)

	}()
}
