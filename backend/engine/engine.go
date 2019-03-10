package engine

import (
	"github.com/anacrolix/torrent"
	"github.com/anatasluo/ant/backend/setting"
	log "github.com/sirupsen/logrus"
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
	engine.EngineRunningInfo = &EngineInfo{}
	engine.EngineRunningInfo.Init()

	engine.setEnvironment()
}

func (engine *Engine)setEnvironment()() {
	err := engine.TorrentDB.DB.One("ID", TorrentLogsID, &engine.EngineRunningInfo.TorrentLogsAndID)
	if err != nil {
		logger.WithFields(log.Fields{"Error":err}).Info("Init running queue now")
	}

	logger.Info("Number of torrent(s) in db is ", len(engine.EngineRunningInfo.TorrentLogs))

	for _, singleLog := range engine.EngineRunningInfo.TorrentLogs {

		if singleLog.Status != DeletedStatus && singleLog.Status != CompletedStatus {
			//fmt.Printf("%s-->%s\n", singleLog.TorrentName, singleLog.Status)
			_, tmpErr := engine.TorrentEngine.AddTorrent(&singleLog.MetaInfo)
			if tmpErr != nil {
				logger.WithFields(log.Fields{"Error":tmpErr}).Info("Failed to add torrent to client")
			}
		}
	}
	engine.EngineRunningInfo.UpdateTorrentLog()
}


func (engine *Engine)Cleanup()() {

	for index := range engine.EngineRunningInfo.TorrentLogs {
		if engine.EngineRunningInfo.TorrentLogs[index].Status != DeletedStatus && engine.EngineRunningInfo.TorrentLogs[index].Status != CompletedStatus {
			if engine.EngineRunningInfo.TorrentLogs[index].Status == RunningStatus {
				engine.StopOneTorrent(engine.EngineRunningInfo.TorrentLogs[index].HashInfoBytes().HexString())
			}
			engine.EngineRunningInfo.TorrentLogs[index].Status = StoppedStatus
		}
	}

	tmpErr := engine.TorrentDB.DB.Save(&engine.EngineRunningInfo.TorrentLogsAndID)
	if tmpErr != nil {
		logger.WithFields(log.Fields{"Error":tmpErr}).Fatal("Failed to save torrent queues")
	}

	engine.TorrentEngine.Close()
	engine.TorrentDB.Cleanup()
}




































