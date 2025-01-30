package states

import (
	"context"
	"fmt"
	"log"

	"github.com/jxs1211/eatfat/internal/server"
	"github.com/jxs1211/eatfat/internal/server/db"
	pkglog "github.com/jxs1211/eatfat/pkg/log"
	"github.com/jxs1211/eatfat/pkg/packets"
)

type BrowsingHiscores struct {
	client  server.ClientInterfacer
	logger  *log.Logger
	queries *db.Queries
	dbCtx   context.Context
}

func (b *BrowsingHiscores) Name() string {
	return "BrowsingHiscores"
}

func (b *BrowsingHiscores) SetClient(client server.ClientInterfacer) {
	b.client = client
	loggingPrefix := fmt.Sprintf("Client %d [%s]: ", client.Id(), b.Name())
	b.logger = pkglog.NewLogger(loggingPrefix)
	b.queries = client.DbTx().Queries
	b.dbCtx = client.DbTx().Ctx
}

func (b *BrowsingHiscores) OnEnter() {
	const limit int64 = 10
	const offset int64 = 0

	topScores, err := b.queries.GetTopScores(b.dbCtx, db.GetTopScoresParams{
		Limit:  limit,
		Offset: offset,
	})
	if err != nil {
		b.logger.Printf("Error getting top %d scores from rank %d: %v", limit, offset+1, err)
		b.client.SocketSend(packets.NewDenyResponse("Failed to get top scores - please try again later"))
		return
	}

	hiscoreMessages := make([]*packets.HiscoreMessage, 0, limit)
	for rank, scoreRow := range topScores {
		hiscoreMessage := &packets.HiscoreMessage{
			Rank:  uint64(rank) + uint64(offset) + 1,
			Name:  scoreRow.Name,
			Score: uint64(scoreRow.BestScore),
		}
		hiscoreMessages = append(hiscoreMessages, hiscoreMessage)
	}

	b.client.SocketSend(packets.NewHiscoreBoard(hiscoreMessages))
}

func (b *BrowsingHiscores) HandleMessage(senderId uint64, message packets.Msg) {
	// handle hiscore exit message
	// b.client.SetState(&Connected{})
}

func (b *BrowsingHiscores) OnExit() {
}
