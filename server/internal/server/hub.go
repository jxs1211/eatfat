package server

import (
	"context"
	"database/sql"
	"log"
	"math/rand/v2"
	"net/http"
	"time"

	_ "embed"

	_ "modernc.org/sqlite"

	"github.com/jxs1211/eatfat/internal/server/collection"
	"github.com/jxs1211/eatfat/internal/server/db"
	"github.com/jxs1211/eatfat/internal/server/objects"
	"github.com/jxs1211/eatfat/pkg/packets"
)

const MaxSpores int = 1000

//go:embed db/config/schema.sql
var schemaGenSql string

// A structure for a state machine to process the client's messages
type ClientStateHandler interface {
	Name() string
	// Inject the client into the state handler
	SetClient(client ClientInterfacer)
	OnEnter()
	HandleMessage(senderId uint64, message packets.Msg)
	// Cleanup the state handler and perform any last actions
	OnExit()
}

// A structure for the connected client to interface with the hub
type ClientInterfacer interface {
	Id() uint64
	ProcessMessage(senderId uint64, message packets.Msg)
	SetState(newState ClientStateHandler)
	// Sets the client's ID and anything else that needs to be initialized
	Initialize(id uint64)
	// Puts data from this client in the write pump
	SocketSend(message packets.Msg)
	// Puts data from another client in the write pump
	SocketSendAs(message packets.Msg, senderId uint64)
	// Forward message to another client for processing
	PassToPeer(message packets.Msg, peerId uint64)
	// Forward message to all other clients for processing
	Broadcast(message packets.Msg)
	// Pump data from the connected socket directly to the client
	ReadPump()
	// Pump data from the client directly to the connected socket
	WritePump()
	// Close the client's connections and cleanup
	Close(reason string)
	DbTx() *DbTx
	SharedGameObjects() *SharedGameObjects
}

// The hub is the central point of communication between all connected clients
type Hub struct {
	Clients           *collection.SharedCollection[ClientInterfacer]
	SharedGameObjects *SharedGameObjects
	// Packets in this channel will be processed by all connected clients except the sender
	BroadcastChan chan *packets.Packet
	// Clients in this channel will be registered with the hub
	RegisterChan chan ClientInterfacer
	// Clients in this channel will be unregistered with the hub
	UnregisterChan chan ClientInterfacer
	// Database connection pool
	dbPool *sql.DB
}

// A structure for database transaction context
type DbTx struct {
	Ctx     context.Context
	Queries *db.Queries
}

func (h *Hub) NewDbTx() *DbTx {
	return &DbTx{
		Ctx:     context.Background(),
		Queries: db.New(h.dbPool),
	}
}

func NewHub() *Hub {
	dbPool, err := sql.Open("sqlite", "db.sqlite")
	if err != nil {
		log.Fatal(err)
	}
	return &Hub{
		Clients: collection.NewSharedCollection[ClientInterfacer](),
		SharedGameObjects: &SharedGameObjects{
			Players: collection.NewSharedCollection[*objects.Player](),
			Spores:  collection.NewSharedCollection[*objects.Spore](),
		},
		BroadcastChan:  make(chan *packets.Packet),
		RegisterChan:   make(chan ClientInterfacer),
		UnregisterChan: make(chan ClientInterfacer),
		dbPool:         dbPool,
	}
}

func (h *Hub) newSpore() *objects.Spore {
	sporeRadius := max(rand.NormFloat64()*3+10, 5)
	x, y := objects.SpawnCoords(sporeRadius, h.SharedGameObjects.Players, h.SharedGameObjects.Spores)
	return &objects.Spore{X: x, Y: y, Radius: sporeRadius}
}

func (h *Hub) Run() {
	if _, err := h.dbPool.ExecContext(context.Background(), schemaGenSql); err != nil {
		log.Fatal(err)
	}
	log.Println("Initializing database done")
	for i := 0; i < MaxSpores; i++ {
		h.SharedGameObjects.Spores.Add(h.newSpore())
	}
	log.Println("Placing spores done")
	go h.replenishSporesLoop(2 * time.Second)
	for {
		select {
		case client := <-h.RegisterChan:
			client.Initialize(h.Clients.Add(client))
		case client := <-h.UnregisterChan:
			h.Clients.Remove(client.Id())
		case packet := <-h.BroadcastChan:
			h.Clients.ForEach(func(id uint64, client ClientInterfacer) {
				if id != packet.SenderId {
					client.ProcessMessage(packet.SenderId, packet.Msg)
				}
			})
		}
	}
}

// Creates a client for the new connection and begins the concurrent read and write pumps
func (h *Hub) Serve(getNewClient func(*Hub, http.ResponseWriter, *http.Request) (ClientInterfacer, error), writer http.ResponseWriter, request *http.Request) {
	log.Println("New client connected from", request.RemoteAddr)
	client, err := getNewClient(h, writer, request)

	if err != nil {
		log.Printf("Error obtaining client for new connection: %v", err)
		return
	}

	h.RegisterChan <- client

	go client.WritePump()
	go client.ReadPump()
}

type SharedGameObjects struct {
	// The ID of the player is the ID of the client
	Players *collection.SharedCollection[*objects.Player]
	Spores  *collection.SharedCollection[*objects.Spore]
}

func (h *Hub) replenishSporesLoop(rate time.Duration) {
	ticker := time.NewTicker(rate)
	defer ticker.Stop()

	for range ticker.C {
		sporesRemaining := h.SharedGameObjects.Spores.Len()
		diff := MaxSpores - sporesRemaining

		if diff <= 0 {
			continue
		}

		log.Printf("%d spores remain - going to replenish %d spores", sporesRemaining, diff)

		// Don't really want to spawn too many at a time, otherwise it can cause a lag spike
		for i := 0; i < min(diff, 10); i++ {
			spore := h.newSpore()
			sporeId := h.SharedGameObjects.Spores.Add(spore)

			h.BroadcastChan <- &packets.Packet{
				SenderId: 0,
				Msg:      packets.NewSpore(sporeId, spore),
			}

			// Sleep a bit to avoid lag spikes
			time.Sleep(50 * time.Millisecond)
		}
	}
}
