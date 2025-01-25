package client

import (
	"context"
	"log"

	"github.com/gorilla/websocket"
	"google.golang.org/protobuf/proto"

	pb "github.com/jxs1211/eatfat/pkg/packets"
)

type Client struct {
	conn *websocket.Conn
}

func NewClient() *Client {
	return &Client{}
}

func (c *Client) Connect(url string) error {
	conn, _, err := websocket.DefaultDialer.Dial(url, nil)
	if err != nil {
		return err
	}
	c.conn = conn
	return nil
}

func (c *Client) SendPacket(packet *pb.Packet) error {
	data, err := proto.Marshal(packet)
	if err != nil {
		return err
	}
	return c.conn.WriteMessage(websocket.BinaryMessage, data)
}

func (c *Client) HandleMessages(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		default:
			messageType, message, err := c.conn.ReadMessage()
			if err != nil {
				log.Printf("Error reading message: %v", err)
				return
			}

			if messageType != websocket.BinaryMessage {
				log.Printf("Received non-binary message type: %d", messageType)
				continue
			}

			packet := &pb.Packet{}
			if err := proto.Unmarshal(message, packet); err != nil {
				log.Printf("Error unmarshaling packet: %v", err)
				continue
			}

			c.handlePacket(packet)
		}
	}
}

func (c *Client) handlePacket(packet *pb.Packet) {
	log.Printf("Received packet from server: %v", packet)
}

func (c *Client) Close() {
	if c.conn != nil {
		c.conn.Close()
	}
}
