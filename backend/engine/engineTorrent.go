package engine

import "github.com/anacrolix/torrent"

func (engine *Engine)AddOneTorrent(filepathAbs string)(tmpTorrent *torrent.Torrent, err error) {
	tmpTorrent, err = engine.TorrentEngine.AddTorrentFromFile(filepathAbs)
	tmpTorrent.DownloadAll()
	return tmpTorrent, err
}
