package engine

import (
	"fmt"
	"github.com/anacrolix/torrent"
	"github.com/anacrolix/torrent/metainfo"
	log "github.com/sirupsen/logrus"
	"time"
)

func (engine *Engine)AddOneTorrent(filepathAbs string)(tmpTorrent *torrent.Torrent, err error) {
	torrentMetaInfo, err := metainfo.LoadFromFile(filepathAbs)
	torrentLog, isExist := engine.EngineRunningInfo.HashToTorrentLog[torrentMetaInfo.HashInfoBytes()]
	if isExist && torrentLog.Status != DeletedStatus && torrentLog.Status != CompletedStatus {
		logger.Info("Task has been created")
		tmpTorrent, _ = engine.TorrentEngine.Torrent(torrentMetaInfo.HashInfoBytes())
		err = nil
	}else if isExist && (torrentLog.Status == DeletedStatus || torrentLog.Status == CompletedStatus) {
		//TODO I can not figure out if user has deleted the file, so I juse see completeStatus as DeletedStatus
		logger.Info("Task has been deleted or completed!We will try to restart it again")
		tmpTorrent, err = engine.TorrentEngine.AddTorrent(torrentMetaInfo)
		torrentLog = engine.EngineRunningInfo.AddoneTorrent(tmpTorrent)
		torrentLog.Status = QueuedStatus
	}else{
		tmpTorrent, err = engine.TorrentEngine.AddTorrent(torrentMetaInfo)
		torrentLog = engine.EngineRunningInfo.AddoneTorrent(tmpTorrent)
	}

	return tmpTorrent, err
}

func (engine *Engine)GetOneTorrent(hexString string)(tmpTorrent *torrent.Torrent, isExist bool) {
	torrentHash := metainfo.Hash{}
	err := torrentHash.FromHexString(hexString)
	if err != nil {
		logger.WithFields(log.Fields{"Error":err}).Error("Unable to get hash from hex string")
	}
	singleTorrentLog, isExist := engine.EngineRunningInfo.HashToTorrentLog[torrentHash]
	if isExist && singleTorrentLog.Status != DeletedStatus && singleTorrentLog.Status != CompletedStatus {
		tmpTorrent, _ = engine.TorrentEngine.Torrent(torrentHash)
	} else {
		tmpTorrent = nil
		isExist = false
	}
	return
}

//TODO: Consider max number of downloading torrents
func (engine *Engine)StartDownloadTorrent(hexString string)(downloaded bool) {
	singleTorrent, isExist := engine.GetOneTorrent(hexString)
	if(isExist) {
		singleTorrentLog, _ := engine.EngineRunningInfo.HashToTorrentLog[singleTorrent.InfoHash()]
		singleTorrentLog.Status = RunningStatus
		engine.EngineRunningInfo.TorrentLogExtends[singleTorrent.InfoHash()] = TorrentLogExtend{
			StatusPub:singleTorrent.SubscribePieceStateChanges(),
			HasStatusPub:true,
		}
		singleTorrent.SetMaxEstablishedConns(clientConfig.EngineSetting.MaxEstablishedConns)
		downloaded = true
		singleTorrent.DownloadAll()
		engine.WaitForCompleted(singleTorrent)
	}else{
		downloaded = false
	}
	return
}

func (engine *Engine)WaitForCompleted(singleTorrent *torrent.Torrent)(){
	go func() {
		singleTorrentLog, _ := engine.EngineRunningInfo.HashToTorrentLog[singleTorrent.InfoHash()]
		singleTorrentLogExtend, _ := engine.EngineRunningInfo.TorrentLogExtends[singleTorrent.InfoHash()]
		engine.ShowTorrentInfo(singleTorrent)
		for singleTorrentLog.Status == RunningStatus {
			if singleTorrent.BytesCompleted() == singleTorrent.Info().TotalLength() {
				log.Info("Task has been finished!")
				singleTorrentLog.Status = CompletedStatus
				singleTorrent.Drop()
				return
			}
			<-singleTorrentLogExtend.StatusPub.Values
			//fmt.Printf("%+v\n", t)
		}
		log.WithFields(log.Fields{"TorrentName":singleTorrentLog.TorrentName, "Status":singleTorrentLog.Status}).Info("Torrent status changed !")
	}()
}

func (engine *Engine)StopOneTorrent(hexString string)(stopped bool) {
	singleTorrent, torrentExist := engine.GetOneTorrent(hexString)
	if torrentExist {
		singleTorrentLog, _:= engine.EngineRunningInfo.HashToTorrentLog[singleTorrent.InfoHash()]
		singleTorrentLog.Status = StoppedStatus
		singleTorrentLogExtend, extendExist := engine.EngineRunningInfo.TorrentLogExtends[singleTorrent.InfoHash()]
		if extendExist && singleTorrentLogExtend.HasStatusPub {
			singleTorrentLogExtend.StatusPub.Values <- log.Fields{"Info":"It should be stopped"}
			singleTorrentLogExtend.HasStatusPub = false
			singleTorrentLogExtend.StatusPub.Close()
		}

		singleTorrent.SetMaxEstablishedConns(0)
		stopped = true
	}else{
		stopped = false
	}
	return
}

func (engine *Engine)ShowTorrentInfo(singleTorrent *torrent.Torrent)  {
	go func() {
		singleTorrentLog, _ := engine.EngineRunningInfo.HashToTorrentLog[singleTorrent.InfoHash()]
		<- singleTorrent.GotInfo()
		for singleTorrentLog.Status == RunningStatus {
			fmt.Printf("%s --> %d \n", singleTorrent.Name(), singleTorrent.BytesCompleted())
			time.Sleep(time.Second)
		}
		log.WithFields(log.Fields{"TorrentName":singleTorrentLog.TorrentName, "Status":singleTorrentLog.Status}).Info("Torrent status changed !")
	}()
}

func (engine *Engine)DelOneTorrent(hexString string)(deleted bool) {
	singleTorrent, torrentExist := engine.GetOneTorrent(hexString)
	if torrentExist {
		singleTorrent.Drop()
		for index := range engine.EngineRunningInfo.TorrentLogs {
			if engine.EngineRunningInfo.TorrentLogs[index].HashInfoBytes().HexString() == hexString {
				engine.EngineRunningInfo.TorrentLogs[index].Status = DeletedStatus
				deleted = true
				break
			}
		}
		engine.EngineRunningInfo.UpdateTorrentLog()
		if !deleted {
			logger.Fatal("Not find deleted hash in logs")
		}
	}else{
		deleted = false
	}
	return
}


