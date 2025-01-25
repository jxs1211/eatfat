package main

import (
	"context"
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/jxs1211/eatfat/pkg/client"
	"github.com/jxs1211/eatfat/pkg/packets"

	pb "github.com/jxs1211/eatfat/pkg/packets"
)

func main() {
	serverAddr := flag.String("server", "ws://127.0.0.1:8080/ws", "WebSocket server address")
	flag.Parse()

	client := client.NewClient()
	log.Printf("Connecting to server %s...", *serverAddr)

	if err := client.Connect(*serverAddr); err != nil {
		log.Fatalf("Failed to connect: %v", err)
	}
	defer client.Close()

	log.Println("Connected to server")

	// Send initial chat message
	packet := &pb.Packet{
		Msg: packets.NewChat("hello"),
	}
	// Handle incoming messages in a goroutine
	ctx, cancel := context.WithCancel(context.Background())
	go client.HandleMessages(ctx)

	if err := client.SendPacket(packet); err != nil {
		log.Printf("Error sending packet: %v", err)
	} else {
		log.Println("Sent packet")
	}

	// Wait for interrupt signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	// Cleanup
	cancel()
	log.Println("Connection closed")
}
