package states

import (
	"fmt"
	"log"

	"github.com/jxs1211/eatfat/internal/server"
	"github.com/jxs1211/eatfat/pkg/packets"
)

type Connected struct {
	client server.ClientInterfacer
	logger *log.Logger
}

func (c *Connected) Name() string {
	return "Connected"
}

func (c *Connected) SetClient(client server.ClientInterfacer) {
	c.client = client
	loggingPrefix := fmt.Sprintf("Client %d [%s]: ", client.Id(), c.Name())
	c.logger = log.New(log.Writer(), loggingPrefix, log.LstdFlags)
}

func (c *Connected) OnEnter() {
	// A newly connected client will want to know its own ID first
	c.client.SocketSend(packets.NewId(c.client.Id()))
}

func (c *Connected) HandleMessage(senderId uint64, message packets.Msg) {
	c.logger.Printf("Received message: %T from client - echoing back...", message)
	if senderId == c.client.Id() {
		// This message was sent by our own client, so broadcast it to everyone else
		c.client.Broadcast(message)
	} else {
		// Another client interfacer passed this onto us, or it was broadcast from the hub,
		// so forward it directly to our own client
		c.client.SocketSendAs(message, senderId)
	}
}

func (c *Connected) OnExit() {
}
