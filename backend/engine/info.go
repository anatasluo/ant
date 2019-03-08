package engine

import (
	"github.com/anacrolix/torrent"
	"github.com/anacrolix/torrent/metainfo"
	"time"
)

type WebviewInfo struct {
	TQueues	TorrentQueues
}

type EngineInfo struct {

}

type TorrentQueues struct {
	ID				int `storm:"id"`
	AllTorrents		map[string]string

}

type TorrentFileList struct {
	TotalFiles 		int				`json:"TotalFiles"`
	FileList		[]TorrentFile	`json:"FileList"`
}

type TorrentList struct {
	NumberOfTorrents	int				`json:"Number"`
	AllTorrent		[]TorrentInfo	`json:"Data"`
}

type PeerFileList struct {
	TotalPeers		int
	PeerList		[]torrent.Peer
}

type TorrentFile struct {
	TorrentHashString		string
	FileName				string
	FilePath				string
	FileSize				string
	FilePercent				string
	FilePriority			string
}

type TorrentInfo struct {
	TorrentHashString		string
	TorrentName				string
	DownloadedSize			string
	DownloadSpeed			string
	Size					string
	Status					string
	PercentDone				string
	ActivePeers				string
	UploadSpeed				string
	StoragePath				string
	DateAdded				string
	ETA						string
	TorrentLabel			string
	SourceType				string
	KnownSwarm				[]torrent.Peer
	UploadRatio				string
	TotalUploadedSize		string
	TotalUploadedBytes		int64
	downloadSpeedInt		int64
	BytesCompleted			int64
	DataBytesWritted		int64
	DataBytesRead			int64
	UpdatedAt				time.Time
	TorrentHash				metainfo.Hash
	NumberOfFiles			int
	NumberOfPieces			int
	MaxConnections			int
}







