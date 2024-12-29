package torrent

import (
	"github.com/negarciacamilo/gorrent/client"
	"github.com/negarciacamilo/gorrent/logger"
	"github.com/negarciacamilo/gorrent/message"
	"github.com/negarciacamilo/gorrent/peer"
	"go.uber.org/zap"
)

type Torrent struct {
	Peers       []peer.Peer
	PeerID      [20]byte
	InfoHash    [20]byte
	PieceHashes [][20]byte
	PieceLength int
	Length      int
	Name        string
}

func (t *Torrent) Download() {
	logger.Info("starting download", zap.String("name", t.Name))

	for _, peer := range t.Peers {
		c, err := client.New(&peer, t.InfoHash, t.PeerID)
		if err != nil {
			break
		}

		defer c.Conn.Close()
		logger.Info("completed handshake", zap.Any("peer", peer.GetFullAddress()))

		c.SendMessage(message.Unchoke)
		c.SendMessage(message.Interested)

		c.Bitfield.HasPiece(0)
	}
}
