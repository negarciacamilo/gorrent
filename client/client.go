package client

import (
	"bytes"
	"errors"
	"github.com/negarciacamilo/gorrent/bitfield"
	"github.com/negarciacamilo/gorrent/handshake"
	"github.com/negarciacamilo/gorrent/logger"
	"github.com/negarciacamilo/gorrent/message"
	"github.com/negarciacamilo/gorrent/peer"
	"go.uber.org/zap"
	"net"
	"time"
)

type Client struct {
	Conn     net.Conn
	Choked   bool
	Bitfield bitfield.Bitfield
	peer     peer.Peer
	infoHash [20]byte
	peerID   [20]byte
}

func New(p *peer.Peer, infoHash, peerId [20]byte) (*Client, error) {
	conn, err := net.DialTimeout("tcp", p.GetFullAddress(), 3*time.Second)
	if err != nil {
		logger.Error("error establishing connection with peer", zap.Any("peer", p.GetFullAddress()))
		return nil, err
	}

	logger.Info("connection established with peer", zap.Any("peer", p.GetFullAddress()))

	_, err = performHandshake(conn, infoHash, peerId)
	if err != nil {
		return nil, err
	}

	// The bitfield message may only be sent immediately after the handshaking sequence is completed, and before any other messages are sent. It is optional, and need not be sent if a client has no pieces.
	bitfield, err := receiveBitfield(conn)
	if err != nil {
		return nil, err
	}

	return &Client{
		Conn:     conn,
		Choked:   true,
		Bitfield: bitfield,
		peer:     *p,
		infoHash: infoHash,
		peerID:   peerId,
	}, nil
}

func performHandshake(conn net.Conn, infoHash, peerId [20]byte) (*handshake.Handshake, error) {
	conn.SetDeadline(time.Now().Add(3 * time.Second))
	defer conn.SetDeadline(time.Time{})

	h := handshake.Handshake{
		PSTR:     "BitTorrent protocol",
		InfoHash: infoHash,
		PeerID:   peerId,
	}
	_, err := conn.Write(h.Serialize())
	if err != nil {
		logger.Error("something happened while trying to handshake peer", zap.Error(err))
		return nil, err
	}

	res, err := handshake.Read(conn)
	if err != nil {
		logger.Error("something happened trying to read peer handshake", zap.Error(err))
		return nil, err
	}

	if !bytes.Equal(res.InfoHash[:], infoHash[:]) {
		msg := "peer infohash doesn't match torrent infohash"
		logger.Error(msg)
		return nil, errors.New(msg)
	}

	return res, nil
}

func receiveBitfield(conn net.Conn) (bitfield.Bitfield, error) {
	conn.SetDeadline(time.Now().Add(5 * time.Second))
	defer conn.SetDeadline(time.Time{})

	msg, err := message.Read(conn)
	if err != nil {
		return nil, err
	}

	if msg == nil {
		errMsg := "Empty message"
		logger.Error(errMsg)
		return nil, errors.New(errMsg)
	}

	if msg.ID != message.Bitfield {
		errMsg := "Unexpected ID, was expecting bitfield"
		logger.Error(errMsg, zap.Any("received-id", msg.ID))
		return nil, errors.New(errMsg)
	}

	return msg.Payload, nil
}

func (c *Client) SendMessage(id message.ID) error {
	msg := message.Message{ID: id}
	_, err := c.Conn.Write(msg.Serialize())
	if err != nil {
		logger.Error("error sending message to peer", zap.Any("message", id), zap.Any("peer", c.peer.GetFullAddress()), zap.Error(err))
		return err
	}
	return nil
}
