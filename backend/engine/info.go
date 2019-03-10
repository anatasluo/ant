package engine

import (
	"github.com/anacrolix/missinggo/pubsub"
	"github.com/anacrolix/torrent"
	"github.com/anacrolix/torrent/metainfo"
)

type WebviewInfo struct {

}

type EngineInfo struct {
	TorrentLogsAndID
	HashToTorrentLog    map[metainfo.Hash]*TorrentLog
	StatusToTorrentLogs map[TorrentStatus][]*TorrentLog
	TorrentLogExtends	map[metainfo.Hash]TorrentLogExtend
}

//These information is needed in running time
type TorrentLogExtend struct {
	StatusPub			*pubsub.Subscription
	HasStatusPub		bool
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
	engineInfo.TorrentLogExtends = make(map[metainfo.Hash]TorrentLogExtend)
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