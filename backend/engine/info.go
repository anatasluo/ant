package engine

import (
	"github.com/anacrolix/missinggo/pubsub"
	"github.com/anacrolix/torrent"
	"github.com/anacrolix/torrent/metainfo"
	"time"
)

type WebviewInfo struct {
	HashToTorrentWebInfo		map[metainfo.Hash]*TorrentWebInfo
}

type EngineInfo struct {
	TorrentLogsAndID
	HashToTorrentLog    map[metainfo.Hash]*TorrentLog
	StatusToTorrentLogs map[TorrentStatus][]*TorrentLog
	TorrentLogExtends	map[metainfo.Hash]*TorrentLogExtend
}

const maxProgressCache = 100
//These information is needed in running time
type TorrentLogExtend struct {
	StatusPub			*pubsub.Subscription
	HasStatusPub		bool
	ProgressInfo		chan TorrentProgressInfo
	HasProgressInfo		bool
	WebNeed				bool
}

//WebInfo only can be used for show in the website, it is generated from engineInfo
type TorrentWebInfo struct {
	TorrentName			string
	TotalLength			string
	HexString			string
	Status				string
	StoragePath 		string
	Percentage			float64
	Files				[]FileInfo
}

type FileInfo struct {
	Path				string
	Priority			byte
	Size				string
}

type TorrentProgressInfo struct {
	Percentage			float64
	UpdateTime			time.Time
}

//This struct will be saved to storm db, so types of its support is limited
type TorrentLog struct {
	metainfo.MetaInfo
	TorrentName			string
	Status				TorrentStatus
	StoragePath 		string
}

type TorrentLogsAndID struct {
	ID          		OnlyStormID 	`storm:"id"`
	TorrentLogs		 	[]TorrentLog
}

type TorrentStatus int

//StormID cant not be zero
const (
	RunningStatus 		TorrentStatus = iota + 1
	QueuedStatus
	StoppedStatus
	CompletedStatus
	DeletedStatus
)

var StatusIDToName = []string {
	"",
	"Running",
	"Queued",
	"Stopped",
	"Completed",
	"Deleted",
}

type OnlyStormID int

const (
	TorrentLogsID		OnlyStormID = iota + 1
)

const (
	nothingStatus		int = iota
	mannualedKill
	osKill
	successCompleted
)

func (engineInfo *EngineInfo) Init()()  {
	engineInfo.ID				= TorrentLogsID
	engineInfo.HashToTorrentLog = make(map[metainfo.Hash]*TorrentLog)
	engineInfo.StatusToTorrentLogs = make(map[TorrentStatus][]*TorrentLog)
	engineInfo.TorrentLogExtends = make(map[metainfo.Hash]*TorrentLogExtend)
}

func (engineInfo *EngineInfo) AddoneTorrent(singleTorrent *torrent.Torrent)(singleTorrentLog *TorrentLog)  {
	singleTorrentLog, isExist := engineInfo.HashToTorrentLog[singleTorrent.InfoHash()]
	if !isExist {
		singleTorrentLog = createTorrenLogFromTorrent(singleTorrent)
		engineInfo.TorrentLogs = append(engineInfo.TorrentLogs, *singleTorrentLog)
		engineInfo.UpdateTorrentLog()
	}
	return
}

func (engineInfo *EngineInfo) UpdateTorrentLog()()  {
	engineInfo.HashToTorrentLog = make(map[metainfo.Hash]*TorrentLog)
	engineInfo.StatusToTorrentLogs = make(map[TorrentStatus][]*TorrentLog)

	for index, singleTorrentLog := range engineInfo.TorrentLogs {
		engineInfo.HashToTorrentLog[singleTorrentLog.HashInfoBytes()] = &engineInfo.TorrentLogs[index]
		engineInfo.StatusToTorrentLogs[singleTorrentLog.Status] = append(engineInfo.StatusToTorrentLogs[singleTorrentLog.Status], &engineInfo.TorrentLogs[index])
	}
}

func createTorrenLogFromTorrent(singleTorrent *torrent.Torrent) *TorrentLog {
	return &TorrentLog{
		MetaInfo		:	singleTorrent.Metainfo(),
		TorrentName		:	singleTorrent.Name(),
		Status			:	QueuedStatus,
	}
}
