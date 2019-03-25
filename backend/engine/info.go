package engine

import (
	"fmt"
	"github.com/anacrolix/missinggo/pubsub"
	"github.com/anacrolix/torrent"
	"github.com/anacrolix/torrent/metainfo"
	"github.com/dustin/go-humanize"
	"math"
	"path/filepath"
	"time"
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
	DownloadSpeed		string
	LeftTime			string
	Files				[]FileInfo
	TorrentStatus		torrent.TorrentStats
	UpdateTime			time.Time
}

type MessageTypeID	int
type MessageFromWeb struct {
	MessageType			MessageTypeID
	HexString			string
}

const (
	updateDuration	    = time.Second
)

const (
	GetInfo				MessageTypeID = iota
)

type FileInfo struct {
	Path				string
	Priority			byte
	Size				string
}

type TorrentProgressInfo struct {
	Percentage			float64
	DownloadSpeed		string
	LeftTime			string
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


func (engineInfo *EngineInfo) init()()  {
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
		logger.Error("Unable to get abs path -> ", err)
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
	<- singleTorrent.GotInfo()
	torrentWebInfo, isExist := engine.WebInfo.HashToTorrentWebInfo[singleTorrent.InfoHash()]
	if !isExist {
		torrentLog, _ := engine.EngineRunningInfo.HashToTorrentLog[singleTorrent.InfoHash()]
		torrentWebInfo = &TorrentWebInfo{
			TorrentName		:	singleTorrent.Info().Name,
			TotalLength		:	generateByteSize(singleTorrent.Info().TotalLength()),
			HexString		:	torrentLog.HashInfoBytes().HexString(),
			Status			:	StatusIDToName[torrentLog.Status],
			StoragePath		:	torrentLog.StoragePath,
			Percentage  	:	float64(singleTorrent.BytesCompleted()) / float64(singleTorrent.Info().TotalLength()),
			DownloadSpeed	:	"Estimating",
			LeftTime		:	"Estimating",
			TorrentStatus	:	singleTorrent.Stats(),
			UpdateTime		:	time.Now(),
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
		torrentWebInfo.TorrentStatus = singleTorrent.Stats()
		torrentWebInfo.Status = StatusIDToName[torrentLog.Status]

		timeNow := time.Now()
		timeDis := timeNow.Sub(torrentWebInfo.UpdateTime).Seconds()
		percentageNow := float64(singleTorrent.BytesCompleted()) / float64(singleTorrent.Info().TotalLength())
		percentageDis := percentageNow - torrentWebInfo.Percentage
		if timeDis >= updateDuration.Seconds() && percentageDis > 0 {
			leftDuration := time.Duration((1 - torrentWebInfo.Percentage) / (percentageDis / timeDis) * 1000 * 1000 * 1000)
			torrentWebInfo.LeftTime = humanizeDuration(leftDuration)
			torrentWebInfo.DownloadSpeed = generateByteSize(int64(percentageDis * float64(singleTorrent.Info().TotalLength()) / timeDis)) + "/s"
			torrentWebInfo.Percentage = percentageNow
			torrentWebInfo.UpdateTime = time.Now()
			if torrentWebInfo.Percentage == 1 {
				engine.CompleteOneTorrent(singleTorrent)
			}
		}
	}
	return
}


func humanizeDuration(duration time.Duration) string {
	if duration.Seconds() < 60.0 {
		return fmt.Sprintf("%d seconds", int64(duration.Seconds()))
	}
	if duration.Minutes() < 60.0 {
		remainingSeconds := math.Mod(duration.Seconds(), 60)
		return fmt.Sprintf("%d minutes %d seconds", int64(duration.Minutes()), int64(remainingSeconds))
	}
	if duration.Hours() < 24.0 {
		remainingMinutes := math.Mod(duration.Minutes(), 60)
		remainingSeconds := math.Mod(duration.Seconds(), 60)
		return fmt.Sprintf("%d hours %d minutes %d seconds",
			int64(duration.Hours()), int64(remainingMinutes), int64(remainingSeconds))
	}
	remainingHours := math.Mod(duration.Hours(), 24)
	remainingMinutes := math.Mod(duration.Minutes(), 60)
	remainingSeconds := math.Mod(duration.Seconds(), 60)
	return fmt.Sprintf("%d days %d hours %d minutes %d seconds",
		int64(duration.Hours()/24), int64(remainingHours),
		int64(remainingMinutes), int64(remainingSeconds))
}