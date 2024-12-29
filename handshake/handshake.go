package handshake

import (
	"github.com/negarciacamilo/gorrent/logger"
	"go.uber.org/zap"
	"io"
)

type Handshake struct {
	PSTR     string
	InfoHash [20]byte
	PeerID   [20]byte
}

func (h *Handshake) Serialize() []byte {
	buf := make([]byte, len(h.PSTR)+49)
	// This is the length of "BitTorrent protocol"
	buf[0] = byte(len(h.PSTR))
	curr := 1
	// Then we send the plain "BitTorrent protocol"
	curr += copy(buf[curr:], h.PSTR)
	// 8 unused bytes for extensions
	curr += copy(buf[curr:], make([]byte, 8))
	// The info hash
	curr += copy(buf[curr:], h.InfoHash[:])
	// And lastly the peer id
	curr += copy(buf[curr:], h.PeerID[:])
	return buf
}

func Read(r io.Reader) (*Handshake, error) {
	lengthBuf := make([]byte, 1)
	_, err := io.ReadFull(r, lengthBuf)
	if err != nil {
		return nil, err
	}
	pstrlen := int(lengthBuf[0])

	if pstrlen == 0 {
		msg := "pstrlen cannot be 0"
		logger.Error(msg)
		return nil, err
	}

	handshakeBuf := make([]byte, 48+pstrlen)
	_, err = io.ReadFull(r, handshakeBuf)
	if err != nil {
		logger.Error("error reading handshake", zap.Error(err))
		return nil, err
	}

	var infoHash, peerID [20]byte

	copy(infoHash[:], handshakeBuf[pstrlen+8:pstrlen+8+20])
	copy(peerID[:], handshakeBuf[pstrlen+8+20:])

	h := Handshake{
		PSTR:     string(handshakeBuf[0:pstrlen]),
		InfoHash: infoHash,
		PeerID:   peerID,
	}

	return &h, nil
}
