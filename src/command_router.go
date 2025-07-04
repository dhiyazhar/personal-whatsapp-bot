package main

import (
	"fmt"
	"strings"

	"go.mau.fi/whatsmeow"
	"go.mau.fi/whatsmeow/types/events"
)

type CommandFunc func(client *whatsmeow.Client, msg *events.Message)

type CommandRouter struct {
	client   *whatsmeow.Client
	commands map[string]CommandFunc
}

func NewCommandRouter(client *whatsmeow.Client) *CommandRouter {
	return &CommandRouter{
		client:   client,
		commands: make(map[string]CommandFunc),
	}
}

func (r *CommandRouter) Register(command string, handler CommandFunc) {
	r.commands[command] = handler
	fmt.Printf("[INFO] command %s berhasil ditambahkan\n", command)
}

func (r *CommandRouter) Handle(msg *events.Message) {
	text := extractTextFromMsg(msg)
	if text == "" {
		fmt.Println("[DEBUG] no text found in message")
	}

	parts := strings.Fields(text)
	if len(parts) == 0 {
		return
	}

	commandKey := strings.ToLower(parts[0])

	handler, ok := r.commands[commandKey]
	if !ok {
		return
	}

	handler(r.client, msg)
}
