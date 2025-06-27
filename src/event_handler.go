package main

import (
	"fmt"

	"go.mau.fi/whatsmeow/types/events"
)

func createEventHandler(router *CommandRouter) func(interface{}) {
	return func(evt interface{}) {
		switch v := evt.(type) {
		case *events.Message:
			router.Handle(v)
		case *events.Connected:
			fmt.Println("[STATUS] Koneksi Berhasi!")
		case *events.Disconnected:
			fmt.Println("[STATUS] Koneksi Terputus.")
		}
	}
}
