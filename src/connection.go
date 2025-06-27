package main

import (
	"context"
	"fmt"
	"os"

	_ "github.com/mattn/go-sqlite3"
	"github.com/mdp/qrterminal/v3"
	"go.mau.fi/whatsmeow"
	"go.mau.fi/whatsmeow/store/sqlstore"
	waLog "go.mau.fi/whatsmeow/util/log"
)

type WhatsAppConnection struct {
	Client *whatsmeow.Client
}

func NewWhatsAppConnection(dbPath string) (*WhatsAppConnection, error) {
	dbLog := waLog.Stdout("Client", "INFO", true)
	ctx := context.Background()

	container, err := sqlstore.New(ctx, "sqlite3", fmt.Sprintf("file:%s?_foreign_keys=on", dbPath), dbLog)
	if err != nil {
		panic(err)
	}

	deviceStore, err := container.GetFirstDevice(ctx)
	if err != nil {
		panic(err)
	}

	clientLog := waLog.Stdout("Client", "INFO", true)
	client := whatsmeow.NewClient(deviceStore, clientLog)

	return &WhatsAppConnection{
		Client: client,
	}, nil

}

func (wac *WhatsAppConnection) Connect() error {
	if wac.Client.Store.ID == nil {
		qrChan, _ := wac.Client.GetQRChannel(context.Background())
		err := wac.Client.Connect()
		if err != nil {
			fmt.Println("gagal menyambungkan client: %w", err)
			panic(err)
		}

		for evt := range qrChan {
			if evt.Event == "code" {
				fmt.Println("-------------------------------------------------")
				fmt.Println("Scan QR code di bawah ini dengan WhatsApp Anda:")
				qrterminal.GenerateHalfBlock(evt.Code, qrterminal.L, os.Stdout)
				fmt.Println("-------------------------------------------------")
			} else {
				fmt.Println("Login event:", evt.Event)
			}
		}

	} else {
		err := wac.Client.Connect()
		if err != nil {
			fmt.Println("gagal menyambungkan kembali client: %w", err)
			panic(err)
		}
	}

	return nil
}

func (wac *WhatsAppConnection) RegisterEventHandler(handler func(evt interface{})) {
	wac.Client.AddEventHandler(handler)
}

func (wac *WhatsAppConnection) Disconnect() {
	fmt.Println("Menutup koneksi...")
	wac.Client.Disconnect()
}
