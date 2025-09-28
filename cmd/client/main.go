package main

import (
	"log"
	"net/http"

	"github.com/TarunAga/adaptive-bitrate-streaming/pkg/server"
)

func main() {
	srv := server.New("./videos") // serve video segments
	log.Println("Starting server on :8080...")
	log.Fatal(http.ListenAndServe(":8080", srv))
}
