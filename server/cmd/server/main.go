package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"

	"github.com/jxs1211/eatfat/internal/server"
	"github.com/jxs1211/eatfat/internal/server/clients"
)

var (
	port = flag.Int("port", 8080, "Port to listen on")
)

func main() {
	flag.Parse()

	hub := server.NewHub()

	// Basic HTTP handler for testing
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "EatFat MMO Server is running!")
	})

	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		hub.Serve(clients.NewWebSocketClient, w, r)
	})

	go hub.Run()
	log.Println("Starting EatFat MMO server on", *port)
	if err := http.ListenAndServe(fmt.Sprintf(":%d", *port), nil); err != nil {
		log.Fatal("Server failed to start:", err)
	}
}
