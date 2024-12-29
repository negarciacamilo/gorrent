package peer

import (
	"crypto/rand"
	"fmt"
	"github.com/negarciacamilo/gorrent/logger"
	"go.uber.org/zap"
)

func GeneratePeerID() [20]byte {
	var uuid [20]byte
	_, err := rand.Read(uuid[:])
	if err != nil {
		// Best effort
		logger.Error("can't generate peer id", zap.Error(err))
	}
	return uuid
}

type Peer struct {
	IP   string `bencode:"ip"`
	Port uint16 `bencode:"port"`
}

func (p *Peer) GetFullAddress() string {
	return fmt.Sprintf("%s:%d", p.IP, p.Port)
}
