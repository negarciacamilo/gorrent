package torrentfile

import (
	"crypto/sha1"
	"errors"
	"github.com/jackpal/bencode-go"
	"github.com/negarciacamilo/gorrent/logger"
	"github.com/negarciacamilo/gorrent/peer"
	torrent "github.com/negarciacamilo/gorrent/torrent"
	"go.uber.org/zap"
	"go.uber.org/zap/buffer"
	"os"
)

const Port = 6881

// This file will handle all the .torrent file operations

type TorrentFile struct {
	// Announce is the URL of the tracker
	Announce string `bencode:"announce"`
	Info     Info   `bencode:"info"`
}

type Info struct {
	// Name maps to a UTF-8 encoded string which is the suggested name to save the file (or directory) as
	Name string `bencode:"name"`
	// PieceLength maps to the number of bytes in each piece the file is split into
	PieceLength int `bencode:"piece length"`
	// Pieces  maps to a string whose length is a multiple of 20
	Pieces string `bencode:"pieces"`
	// Length is the length of the file, in bytes
	Length int `bencode:"length"`
}

// Hash will generate info_hash
func (i *Info) Hash() ([20]byte, error) {
	var buf buffer.Buffer
	err := bencode.Marshal(&buf, *i)
	if err != nil {
		logger.Error("can't generate info_hash", zap.Error(err))
		return [20]byte{}, err
	}

	return sha1.Sum(buf.Bytes()), nil
}

// Pieces are 20 byte long each, so we should split them
func (i *Info) SplitPieces() ([][20]byte, error) {
	hashLength := 20
	buf := []byte(i.Pieces)

	if len(buf)%hashLength != 0 {
		msg := "Malformed pieces"
		logger.Error(msg, zap.Any("Length", len(buf)))
		return nil, errors.New(msg)
	}

	numberOfHashes := len(buf) / hashLength
	hashes := make([][20]byte, numberOfHashes)

	for i := 0; i < numberOfHashes; i++ {
		// When 0, 0 * 20 = 0 : 1 * 20 -> [0:20]
		// 1, 1 * 20 = 20 : 2 * 20 -> [20:40]
		copy(hashes[i][:], buf[i*hashLength:(i+1)*hashLength])
	}

	return hashes, nil
}

func OpenFile(path string) (*TorrentFile, error) {
	f, err := os.Open(path)
	if err != nil {
		logger.Error("can't open torrent file", zap.Error(err))
		return nil, err
	}
	defer f.Close()

	t := TorrentFile{}
	err = bencode.Unmarshal(f, &t)
	if err != nil {
		logger.Error("can't unmarshal torrent file", zap.Error(err), zap.Any("input", f))
		return nil, err
	}

	return &t, nil
}

func (t *TorrentFile) Download() error {
	peerId := peer.GeneratePeerID()

	trackerResponse, err := t.TrackerRequest(peerId, Port)
	if err != nil {
		return err
	}

	infoHash, err := t.Info.Hash()
	if err != nil {
		return err
	}

	splittedPieces, err := t.Info.SplitPieces()
	if err != nil {
		return err
	}

	torrent := torrent.Torrent{
		Peers:       trackerResponse.Peers,
		PeerID:      peerId,
		InfoHash:    infoHash,
		PieceHashes: splittedPieces,
		PieceLength: t.Info.PieceLength,
		Length:      t.Info.Length,
		Name:        t.Info.Name,
	}

	torrent.Download()

	return err
}
