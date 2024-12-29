package main

import (
	"github.com/negarciacamilo/gorrent/logger"
	"github.com/negarciacamilo/gorrent/torrentfile"
)

func main() {
	fileName := "debian.torrent"

	tf, err := torrentfile.OpenFile(fileName)
	if err != nil {
		logger.Panic(err.Error())
	}

	err = tf.Download()
	if err != nil {
		logger.Panic(err.Error())
	}
}
