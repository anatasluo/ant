package engine

import (
	"github.com/anacrolix/torrent"
	"github.com/anacrolix/torrent/metainfo"
	"github.com/anatasluo/ant/backend/setting"
	log "github.com/sirupsen/logrus"
	"path/filepath"
)

type Engine struct {
	TorrentEngine		*torrent.Client
	TorrentDB			*TorrentDB
	WebInfo				*WebviewInfo
	EngineRunningInfo	*EngineInfo
}

var (
	onlyEngine				Engine
	hasCreatedEngine	= 	false
	clientConfig 		=	setting.GetClientSetting()
	logger 				=	clientConfig.LoggerSetting.Logger
)

func GetEngine() *Engine {
	if hasCreatedEngine == false {
		onlyEngine.initAndRunEngine()
		hasCreatedEngine = true
	}
	return &onlyEngine
}

func (engine *Engine)initAndRunEngine()()  {
	engine.TorrentDB = GetTorrentDB(clientConfig.EngineSetting.TorrentDBPath)

	var tmpErr error
	engine.TorrentEngine, tmpErr = torrent.NewClient(&clientConfig.EngineSetting.TorrentConfig)
	if tmpErr != nil {
		logger.WithFields(log.Fields{"Error":tmpErr}).Error("Failed to Created torrent engine")
	}

	engine.WebInfo = &WebviewInfo{}
	engine.WebInfo.HashToTorrentWebInfo = make(map[metainfo.Hash]*TorrentWebInfo)

	engine.EngineRunningInfo = &EngineInfo{}
	engine.EngineRunningInfo.init()

	//Get info from storm database
	engine.setEnvironment()
}

func (engine *Engine)setEnvironment()() {

	engine.TorrentDB.GetLogs(&engine.EngineRunningInfo.TorrentLogsAndID)

	logger.Debug("Number of torrent(s) in db is ", len(engine.EngineRunningInfo.TorrentLogs))

	for _, singleLog := range engine.EngineRunningInfo.TorrentLogs {

		if singleLog.Status != CompletedStatus {
			_, tmpErr := engine.TorrentEngine.AddTorrent(&singleLog.MetaInfo)
			if tmpErr != nil {
				logger.WithFields(log.Fields{"Error":tmpErr}).Info("Failed to add torrent to client")
			}
		}
	}
	engine.UpdateInfo()
}

func (engine *Engine)Restart()() {
	logger.Info("Restart engine")

	//To handle problems caused by change of settings
	for index := range engine.EngineRunningInfo.TorrentLogs {
		if engine.EngineRunningInfo.TorrentLogs[index].Status != CompletedStatus && engine.EngineRunningInfo.TorrentLogs[index].StoragePath != clientConfig.TorrentConfig.DataDir{
			filePath := filepath.Join(engine.EngineRunningInfo.TorrentLogs[index].StoragePath, engine.EngineRunningInfo.TorrentLogs[index].TorrentName)
			log.WithFields(log.Fields{"Path":filePath}).Info("To restart engine, these unfinished files will be deleted")
			singleTorrent, torrentExist := engine.GetOneTorrent(engine.EngineRunningInfo.TorrentLogs[index].HashInfoBytes().HexString())
			if torrentExist {
				singleTorrent.Drop()
			}
			engine.EngineRunningInfo.TorrentLogs[index].StoragePath = clientConfig.TorrentConfig.DataDir
			engine.UpdateInfo()
			delFiles(filePath)
		}
	}
	engine.Cleanup()
	GetEngine()
}

func (engine *Engine)SaveInfo()() {
	tmpErr := engine.TorrentDB.DB.Save(&engine.EngineRunningInfo.TorrentLogsAndID)
	if tmpErr != nil {
		logger.WithFields(log.Fields{"Error":tmpErr}).Fatal("Failed to save torrent queues")
	}
}

func (engine *Engine)Cleanup()() {

	hasCreatedEngine = false
	engine.UpdateInfo()

	for index := range engine.EngineRunningInfo.TorrentLogs {
		if engine.EngineRunningInfo.TorrentLogs[index].Status != CompletedStatus {
			if engine.EngineRunningInfo.TorrentLogs[index].Status == AnalysingStatus {
				aimLog := engine.EngineRunningInfo.TorrentLogs[index]
				torrentHash := metainfo.Hash{}
				_ = torrentHash.FromHexString(aimLog.TorrentName)
				magnetTorrent, isExist := engine.TorrentEngine.Torrent(torrentHash)
				if isExist {
					logger.Info("One magnet will be deleted " + magnetTorrent.String())
					magnetTorrent.Drop()
				}
			}else if engine.EngineRunningInfo.TorrentLogs[index].Status == RunningStatus {
				engine.StopOneTorrent(engine.EngineRunningInfo.TorrentLogs[index].HashInfoBytes().HexString())
				engine.EngineRunningInfo.TorrentLogs[index].Status = StoppedStatus
			}else if engine.EngineRunningInfo.TorrentLogs[index].Status == QueuedStatus{
				engine.EngineRunningInfo.TorrentLogs[index].Status = StoppedStatus
			}
		}
	}

	//Update info in torrentLogs, remove magnet
	tmpLogs := engine.EngineRunningInfo.TorrentLogs
	engine.EngineRunningInfo.TorrentLogs = nil

	for index := range tmpLogs {
		if tmpLogs[index].Status != AnalysingStatus {
			engine.EngineRunningInfo.TorrentLogs = append(engine.EngineRunningInfo.TorrentLogs, tmpLogs[index])
		}
	}

	engine.SaveInfo()

	engine.TorrentEngine.Close()
	engine.TorrentDB.Cleanup()
}




































