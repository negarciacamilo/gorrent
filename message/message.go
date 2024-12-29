package message

import (
	"encoding/binary"
	"github.com/negarciacamilo/gorrent/logger"
	"go.uber.org/zap"
	"io"
)

type ID uint8

const (
	Choke         ID = 0
	Unchoke       ID = 1
	Interested    ID = 2
	NotInterested ID = 3
	Have          ID = 4
	Bitfield      ID = 5
	Request       ID = 6
	Piece         ID = 7
	Cancel        ID = 8
)

type Message struct {
	ID      ID
	Payload []byte
}

func (m *Message) Serialize() []byte {
	if m == nil {
		return make([]byte, 4)
	}

	length := uint32(len(m.Payload) + 1)
	buf := make([]byte, 4+length)
	binary.BigEndian.PutUint32(buf[0:4], length)
	buf[4] = byte(m.ID)
	copy(buf[5:], m.Payload)
	return buf
}

func Read(r io.Reader) (*Message, error) {
	lengthBuf := make([]byte, 4)
	_, err := io.ReadFull(r, lengthBuf)
	if err != nil {
		logger.Error("error reading length bytes", zap.Error(err))
		return nil, err
	}
	length := binary.BigEndian.Uint32(lengthBuf)

	if length == 0 {
		return nil, nil
	}

	messageBuf := make([]byte, length)
	_, err = io.ReadFull(r, messageBuf)
	if err != nil {
		logger.Error("error reading message bytes", zap.Error(err))
		return nil, err
	}

	m := Message{
		ID:      ID(messageBuf[0]),
		Payload: messageBuf[1:],
	}

	return &m, nil
}
