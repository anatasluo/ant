package engine

import (
	"github.com/anacrolix/torrent"
	"github.com/anatasluo/ant/backend/setting"
	log "github.com/sirupsen/logrus"
	"io/ioutil"
	"os"
	"time"
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
	var tmpErr error
	engine.TorrentEngine, tmpErr = torrent.NewClient(&clientConfig.EngineSetting.TorrentConfig)
	if tmpErr != nil {
		logger.WithFields(log.Fields{"Error":tmpErr}).Error("Failed to Created torrent engine")
	}
	engine.TorrentDB = GetTorrentDB(clientConfig.EngineSetting.TorrentDBPath)

	engine.WebInfo = &WebviewInfo{}
	engine.EngineRunningInfo = &EngineInfo{}

	engine.setEnvironment()
}

func (engine *Engine)setEnvironment()() {
	err := engine.TorrentDB.DB.One("ID", 5, &engine.WebInfo.TQueues)
	if err != nil {
		logger.WithFields(log.Fields{"Error":err}).Info("Init running queue now")
		engine.WebInfo.TQueues.ID = 5
		tmpErr := engine.TorrentDB.DB.Save(&engine.WebInfo.TQueues)
		if tmpErr != nil {
			logger.WithFields(log.Fields{"Error":err}).Info("Failed to save torrent queues")
		}
	}



}

func (engine *Engine)waitForInfoByte(singleTorrent *torrent.Torrent, seconds time.Duration)(downloaded bool) {
	logger.WithFields(log.Fields{"Seconds to wait for info " : seconds}).Info("start to download info for a torrent")
	timeout := make(chan bool, 1)
	go func() {
		time.Sleep(seconds * time.Second)
	}()
	select {
	case <- singleTorrent.GotInfo():
			log.Info("Get info fo torrent successfully")
			return true
	case <- timeout:
			log.Info("Failed to get info for a torrent")
			singleTorrent.Drop()
			return false
	}
}


func (engine *Engine)readTorrentFile(elem *TorrentLocal)(singleTorrent *torrent.Torrent, err error) {
	tempFile, err := ioutil.TempFile("", "TorrentFileTemp")
	if err != nil {
		logger.WithFields(log.Fields{"TempFile": tempFile, "Error": err}).Error("Unable to create a temp file")
		return nil, err
	}

	//TODO
	defer os.Remove(tempFile.Name())

	if _, err := tempFile.Write(elem.TorrentFile); err != nil {
		logger.WithFields(log.Fields{"TempFile": tempFile, "Error": err}).Error("Unable to write a temp file")
		return nil, err
	}

	if err := tempFile.Close(); err != nil {
		logger.WithFields(log.Fields{"TempFile": tempFile, "Error": err}).Error("Unable to close a temp file")
	}

	if _, err = os.Stat(elem.TorrentFileName); err != nil {
		logger.WithFields(log.Fields{"TempFileName": elem.TorrentFileName, "Error": err}).Error("Unable to find torrent file")
		engine.TorrentDB.DelOneTorrent(elem.Hash)
		return nil, err
	}

	if singleTorrent, err = engine.TorrentEngine.AddTorrentFromFile(elem.TorrentFileName); err !=nil {
		logger.WithFields(log.Fields{"TempFileName": elem.TorrentFileName, "Error": err}).Error("Unable to add torrent file")
		engine.TorrentDB.DelOneTorrent(elem.Hash)
		return nil, err
	}

	return
}

func (engine *Engine)Cleanup()() {
	engine.TorrentDB.Cleanup()
}




































