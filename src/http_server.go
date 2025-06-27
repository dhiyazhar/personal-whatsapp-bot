package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
)

func healthCheckHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "Bot is alive and listening!")
}

func startHealthCheckServer() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	http.HandleFunc("/healthz", healthCheckHandler)

	log.Printf("Health check server starting on internal port: %s", port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatalf("Fatal: Failed to start health check server: %v", err)
	}
}
