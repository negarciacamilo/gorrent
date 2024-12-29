package torrentfile

import (
	"bytes"
	"errors"
	"github.com/go-resty/resty/v2"
	"github.com/jackpal/bencode-go"
	"github.com/negarciacamilo/gorrent/logger"
	"github.com/negarciacamilo/gorrent/peer"
	"go.uber.org/zap"
	"net/url"
	"strconv"
)

type TrackerResponse struct {
	// How much should we wait, in seconds
	Interval int `bencode:"interval"`
	// Peers is a list of dictionaries corresponding to peers, each of which contains the keys peer id, ip,
	// and port, which map to the peer's self-selected ID, IP address or dns name as a string, and port number, respectively
	Peers []peer.Peer `bencode:"peers"`
}

func (t *TorrentFile) TrackerRequest(peerId [20]byte, port int) (*TrackerResponse, error) {
	trackerURL, err := buildTrackerURL(t, peerId, port)
	if err != nil {
		return nil, err
	}

	client := resty.New()
	resp, err := client.R().Get(trackerURL)
	if err != nil {
		logger.Error("something happened performing tracker request", zap.Error(err))
		return nil, err
	}

	if resp.IsError() {
		msg := "tracker request not succeeded"
		logger.Error(msg, zap.Any("response", resp.String()), zap.Int("status", resp.StatusCode()))
		return nil, errors.New(msg)
	}

	var trackerResponse TrackerResponse
	err = bencode.Unmarshal(bytes.NewReader(resp.Body()), &trackerResponse)
	if err != nil {
		logger.Error("cannot parse tracker response", zap.Any("response", resp.String()))
		return nil, err
	}

	return &trackerResponse, nil
}

/*
The tracker request contains the following fields:

info_hash
The 20 byte sha1 hash of the bencoded form of the info value from the metainfo file. This value will almost certainly have to be escaped.

peer_id
A string of length 20 which this downloader uses as its id. Each downloader generates its own id at random at the start of a new download. This value will also almost certainly have to be escaped.

ip
An optional parameter giving the IP (or dns name) which this peer is at. Generally used for the origin if it's on the same machine as the tracker.

port
The port number this peer is listening on. Common behavior is for a downloader to try to listen on port 6881 and if that port is taken try 6882, then 6883, etc. and give up after 6889.

uploaded
The total amount uploaded so far, encoded in base ten ascii.

downloaded
The total amount downloaded so far, encoded in base ten ascii.

left
The number of bytes this peer still has to download, encoded in base ten ascii. Note that this can't be computed from downloaded and the file length since it might be a resume, and there's a chance that some of the downloaded data failed an integrity check and had to be re-downloaded.

event
This is an optional key which maps to started, completed, or stopped (or empty, which is the same as not being present). If not present, this is one of the announcements done at regular intervals. An announcement using started is sent when a download first begins, and one using completed is sent when the download is complete. No completed is sent if the file was complete when started. Downloaders send an announcement using stopped when they cease downloading.
*/
func buildTrackerURL(t *TorrentFile, peerId [20]byte, port int) (string, error) {
	parsedURL, err := url.Parse(t.Announce)
	if err != nil {
		logger.Error("can't parse announce url", zap.Error(err), zap.Any("torrent", t))
		return "", err
	}

	infoHash, err := t.Info.Hash()
	if err != nil {
		return "", err
	}

	params := url.Values{
		"info_hash":  []string{string(infoHash[:])},
		"peer_id":    []string{string(peerId[:])},
		"port":       []string{strconv.Itoa(port)},
		"uploaded":   []string{"0"},
		"downloaded": []string{"0"},
		"left":       []string{strconv.Itoa(t.Info.Length)},
	}

	parsedURL.RawQuery = params.Encode()
	return parsedURL.String(), nil
}
