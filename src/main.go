package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	// config, err := LoadConfig()
	// if err != nil {
	// 	fmt.Printf("Failed to load config: %v\n", err)
	// }

	wa, err := NewWhatsAppConnection("whatsapp.db")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error inisialisasi client: %v\n", err)
		os.Exit(1)
	}

	jobsChannel := make(chan DownloadJob, 50)

	go StartDownloadWorker(wa.Client, jobsChannel)

	router := NewCommandRouter(wa.Client, jobsChannel)

	router.Register("!ping", handlePing)
	router.Register("!download", handleDownload)
	router.Register("!d", handleDownload)

	mainHandler := createEventHandler(router)
	wa.RegisterEventHandler(mainHandler)

	err = wa.Connect()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error saat menyambungkan: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Bot berjalan...")
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	<-c

	wa.Disconnect()
}
