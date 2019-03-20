package engine

import (
	"github.com/anacrolix/missinggo/pubsub"
	"github.com/anacrolix/torrent"
	"github.com/anacrolix/torrent/metainfo"
	"github.com/dustin/go-humanize"
	"path/filepath"
)

type WebviewInfo struct {
	HashToTorrentWebInfo		map[metainfo.Hash]*TorrentWebInfo
	//Store magnets that have not get meta data
	MagnetTmpInfo				map[metainfo.Hash]*TorrentWebInfo
}

type EngineInfo struct {
	TorrentLogsAndID
	HashToTorrentLog    map[metainfo.Hash]*TorrentLog
	TorrentLogExtends	map[metainfo.Hash]*TorrentLogExtend
}

//These information is needed in running time
type TorrentLogExtend struct {
	StatusPub			*pubsub.Subscription
	HasStatusPub		bool
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
	TorrentStatus		torrent.TorrentStats
}

type MessageFromWeb struct {
	HexString			string
}

type FileInfo struct {
	Path				string
	Priority			byte
	Size				string
}

type TorrentProgressInfo struct {
	Percentage			float64
	HexString			string
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
	QueuedStatus 		TorrentStatus = iota + 1
	//This status only used for magnet
	AnalysingStatus
	RunningStatus
	StoppedStatus
	CompletedStatus
)

var StatusIDToName = []string {
	"",
	"Queued",
	"Analysing",
	"Running",
	"Stopped",
	"Completed",
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
	engineInfo.TorrentLogExtends = make(map[metainfo.Hash]*TorrentLogExtend)
}

func (engineInfo *EngineInfo) AddOneTorrent(singleTorrent *torrent.Torrent)(singleTorrentLog *TorrentLog)  {
	singleTorrentLog, isExist := engineInfo.HashToTorrentLog[singleTorrent.InfoHash()]
	if !isExist {
		singleTorrentLog = createTorrentLogFromTorrent(singleTorrent)
		engineInfo.TorrentLogs = append(engineInfo.TorrentLogs, *singleTorrentLog)
		engineInfo.UpdateTorrentLog()
	}
	return
}

func (engine *Engine) UpdateInfo()() {
	engine.EngineRunningInfo.UpdateTorrentLog();
	engine.UpdateWebInfo()
}

func (engineInfo *EngineInfo) UpdateTorrentLog()()  {
	engineInfo.HashToTorrentLog = make(map[metainfo.Hash]*TorrentLog)

	for index, singleTorrentLog := range engineInfo.TorrentLogs {
		engineInfo.HashToTorrentLog[singleTorrentLog.HashInfoBytes()] = &engineInfo.TorrentLogs[index]
	}
}

func (engine *Engine) UpdateWebInfo()()  {
	engine.WebInfo.HashToTorrentWebInfo = make(map[metainfo.Hash]*TorrentWebInfo)
	for _, singleTorrent := range engine.TorrentEngine.Torrents() {
		engine.WebInfo.HashToTorrentWebInfo[singleTorrent.InfoHash()] = engine.GenerateInfoFromTorrent(singleTorrent)
	}
}

func createTorrentLogFromTorrent(singleTorrent *torrent.Torrent) *TorrentLog {
	absPath, err := filepath.Abs(clientConfig.EngineSetting.TorrentConfig.DataDir)
	if err != nil {
		logger.Info("Unable to get abs path -> ", err)
	}
	return &TorrentLog{
		MetaInfo		:	singleTorrent.Metainfo(),
		TorrentName		:	singleTorrent.Name(),
		Status			:	QueuedStatus,
		StoragePath		:	absPath,
	}
}

func generateByteSize(byteSize int64) string {
	return humanize.Bytes(uint64(byteSize))
}

//For complete status
func (engine *Engine) GenerateInfoFromLog(torrentLog TorrentLog) (torrentWebInfo *TorrentWebInfo) {
	torrentWebInfo, _ = engine.WebInfo.HashToTorrentWebInfo[torrentLog.HashInfoBytes()]
	torrentWebInfo = &TorrentWebInfo{
		TorrentName		:	torrentLog.TorrentName,
		HexString		:	torrentLog.HashInfoBytes().HexString(),
		Status			:	StatusIDToName[torrentLog.Status],
		StoragePath		:	torrentLog.StoragePath,
		Percentage		:	1,
	}
	engine.WebInfo.HashToTorrentWebInfo[torrentLog.HashInfoBytes()] = torrentWebInfo
	return
}

//For magnet
func (engine *Engine) GenerateInfoFromMagnet(infoHash metainfo.Hash) (torrentWebInfo *TorrentWebInfo) {
	return &TorrentWebInfo{
		TorrentName		: infoHash.String(),
		HexString		: infoHash.String(),
		Status			:  StatusIDToName[AnalysingStatus],
		StoragePath		: "",
		Percentage		:  0,
	}
}

func (engine *Engine) GenerateInfoFromTorrent(singleTorrent *torrent.Torrent) (torrentWebInfo *TorrentWebInfo) {
	torrentWebInfo, isExist := engine.WebInfo.HashToTorrentWebInfo[singleTorrent.InfoHash()]
	if !isExist {
		<- singleTorrent.GotInfo();
		torrentLog, _ := engine.EngineRunningInfo.HashToTorrentLog[singleTorrent.InfoHash()]
		torrentWebInfo = &TorrentWebInfo{
			TorrentName		:	singleTorrent.Info().Name,
			TotalLength		:	generateByteSize(singleTorrent.Info().TotalLength()),
			HexString		:	torrentLog.HashInfoBytes().HexString(),
			Status			:	StatusIDToName[torrentLog.Status],
			StoragePath		:	torrentLog.StoragePath,
			Percentage  	:	float64(singleTorrent.BytesCompleted()) / float64(singleTorrent.Info().TotalLength()),
			TorrentStatus	:	singleTorrent.Stats(),
		}
		for _, key := range singleTorrent.Files() {
			torrentWebInfo.Files = append(torrentWebInfo.Files, FileInfo{
				Path	:	key.Path(),
				Priority:	byte(key.Priority()),
				Size	:	generateByteSize(key.Length()),
			})
		}
		engine.WebInfo.HashToTorrentWebInfo[singleTorrent.InfoHash()] = torrentWebInfo
	}else{
		torrentLog, _ := engine.EngineRunningInfo.HashToTorrentLog[singleTorrent.InfoHash()]
		torrentWebInfo.Status = StatusIDToName[torrentLog.Status]
		torrentWebInfo.Percentage = float64(singleTorrent.BytesCompleted()) / float64(singleTorrent.Info().TotalLength())
		torrentWebInfo.TorrentStatus = singleTorrent.Stats()
	}
	return
}

