package engine

import (
	"errors"
	"fmt"
	"github.com/anacrolix/torrent"
	"github.com/anacrolix/torrent/metainfo"
	log "github.com/sirupsen/logrus"
	"os"
	"path/filepath"
	"strings"
)



func (engine *Engine)AddOneTorrentFromFile (filepathAbs string)(tmpTorrent *torrent.Torrent, err error) {
	torrentMetaInfo, err := metainfo.LoadFromFile(filepathAbs)
	if err == nil {
		//To solve problem of different variable scope
		needMoreOperation := false
		tmpTorrent, needMoreOperation = engine.checkOneHash(torrentMetaInfo.HashInfoBytes())
		if needMoreOperation {
			tmpTorrent, err = engine.TorrentEngine.AddTorrent(torrentMetaInfo)
			engine.EngineRunningInfo.AddOneTorrent(tmpTorrent)
		}
	}
	return tmpTorrent, err
}

//Check if duplicated torrent
func (engine *Engine)checkOneHash(infoHash metainfo.Hash) (tmpTorrent *torrent.Torrent, needMoreOperation bool){
	torrentLog, isExist := engine.EngineRunningInfo.HashToTorrentLog[infoHash]
	if isExist && torrentLog.Status != CompletedStatus {
		logger.Debug("Task has been created")
		tmpTorrent, _ = engine.TorrentEngine.Torrent(infoHash)
		needMoreOperation = false
	}else if isExist && torrentLog.Status == CompletedStatus {
		logger.Debug("Task has been completed!")
		tmpTorrent = nil
		needMoreOperation = false
	}else{
		logger.Debug("It is a new torrent")
		needMoreOperation = true
	}
	return
}

//TODO: There are some problems for client to add magnet directly, so I just use itorrents api to convert magneg to torrent in Electron
func (engine *Engine)AddOneTorrentFromMagnet (linkAddress string)(tmpTorrent *torrent.Torrent, err error) {
	if strings.HasPrefix(linkAddress, "magnet:") {
		torrentMetaInfo, err := torrent.TorrentSpecFromMagnetURI(linkAddress)
		if err != nil {
			logger.WithFields(log.Fields{"Error":err}).Error("unable to resolve magnet")
		}
		var needMoreOperation bool
		tmpTorrent, needMoreOperation = engine.checkOneHash(torrentMetaInfo.InfoHash)
		//To display for webview
		engine.WebInfo.MagnetTmpInfo[torrentMetaInfo.InfoHash] = engine.GenerateInfoFromMagnet(torrentMetaInfo.InfoHash)

		if needMoreOperation {
			go func() {
				tmpTorrent, _, err := engine.TorrentEngine.AddTorrentSpec(torrentMetaInfo)
				if err != nil {
					logger.WithFields(log.Fields{"Error":err, "Torrent": tmpTorrent}).Error("Unable to resolve magnet")
				}
				if tmpTorrent != nil && err != nil {
					logger.Info("Add torrent from magnet")
					delete(engine.WebInfo.MagnetTmpInfo, torrentMetaInfo.InfoHash)
					engine.EngineRunningInfo.AddOneTorrent(tmpTorrent)
					engine.GenerateInfoFromTorrent(tmpTorrent)
					engine.StartDownloadTorrent(tmpTorrent.InfoHash().HexString())
				}
			}()
			err = nil
		}
	} else if strings.HasPrefix(linkAddress, "infohash:") {
		infoHash := metainfo.NewHashFromHex(strings.TrimPrefix(linkAddress, "infohash:"))
		var needMoreOperation bool
		tmpTorrent, needMoreOperation = engine.checkOneHash(infoHash)
		if needMoreOperation {
			tmpTorrent, _ = engine.TorrentEngine.AddTorrentInfoHash(infoHash)
			err = nil
			engine.EngineRunningInfo.AddOneTorrent(tmpTorrent)
		}
	} else {
		tmpTorrent = nil
		err = errors.New("Invalid address")
	}
	return tmpTorrent, err
}

//Only handle torrent in client
func (engine *Engine)GetOneTorrent(hexString string)(tmpTorrent *torrent.Torrent, isExist bool) {
	torrentHash := metainfo.Hash{}
	err := torrentHash.FromHexString(hexString)
	if err != nil {
		logger.WithFields(log.Fields{"Error":err}).Error("Unable to get hash from hex string")
	}
	tmpTorrent, isExist = engine.TorrentEngine.Torrent(torrentHash)
	return
}

//Max number of downloading torrents should be considered in electron
func (engine *Engine)StartDownloadTorrent(hexString string)(downloaded bool) {
	downloaded = true
	singleTorrent, isExist := engine.GetOneTorrent(hexString)
	if isExist {
		singleTorrentLog, _ := engine.EngineRunningInfo.HashToTorrentLog[singleTorrent.InfoHash()]
		if singleTorrentLog.Status != RunningStatus {
			singleTorrentLog.Status = RunningStatus
			_, extendIsExist := engine.EngineRunningInfo.TorrentLogExtends[singleTorrent.InfoHash()];
			if !extendIsExist {
				engine.EngineRunningInfo.TorrentLogExtends[singleTorrent.InfoHash()] = &TorrentLogExtend{
					StatusPub:singleTorrent.SubscribePieceStateChanges(),
					HasStatusPub:true,
				}
			}
			logger.Debug("Create extend info for log")
			//Some download setting for task
			logger.Debug(clientConfig.DefaultTrackers)
			singleTorrent.AddTrackers(clientConfig.DefaultTrackers)
			singleTorrent.SetMaxEstablishedConns(clientConfig.EngineSetting.MaxEstablishedConns)
			engine.WaitForCompleted(singleTorrent)
			//TODO: Download selected files
			singleTorrent.DownloadAll()
		}
	}else{
		downloaded = false
	}
	return
}

func (engine *Engine)CompleteOneTorrent(singleTorrent *torrent.Torrent)() {
	singleTorrentLog, _ := engine.EngineRunningInfo.HashToTorrentLog[singleTorrent.InfoHash()]
	singleTorrentLogExtend, extendExist := engine.EngineRunningInfo.TorrentLogExtends[singleTorrent.InfoHash()]
	<- singleTorrent.GotInfo()
	//One more check
	if singleTorrent.BytesCompleted() == singleTorrent.Info().TotalLength() {
		logger.WithFields(log.Fields{"TorrentName": singleTorrent.Name()}).Info("Torrent has been finished")
		singleTorrent.VerifyData()
		singleTorrentLog.Status = CompletedStatus
		if extendExist && singleTorrentLogExtend.HasStatusPub && singleTorrentLogExtend.StatusPub != nil {
			singleTorrentLogExtend.HasStatusPub = false
			if !channelClosed(singleTorrentLogExtend.StatusPub.Values) {
				singleTorrentLogExtend.StatusPub.Values <- struct{}{}
				singleTorrentLogExtend.StatusPub.Close()
			}
		}
	}
}

func (engine *Engine)WaitForCompleted(singleTorrent *torrent.Torrent)(){
	go func() {
		singleTorrentLog, _ := engine.EngineRunningInfo.HashToTorrentLog[singleTorrent.InfoHash()]
		singleTorrentLogExtend, _ := engine.EngineRunningInfo.TorrentLogExtends[singleTorrent.InfoHash()]
		<- singleTorrent.GotInfo()
		for singleTorrentLog.Status == RunningStatus {
			if singleTorrent.BytesCompleted() == singleTorrent.Info().TotalLength() {
				engine.CompleteOneTorrent(singleTorrent)
				engine.UpdateInfo()
				return
			}
			<-singleTorrentLogExtend.StatusPub.Values
		}
		log.WithFields(log.Fields{"TorrentName":singleTorrentLog.TorrentName, "Status":singleTorrentLog.Status}).Info("Torrent status changed !")
	}()
}

func (engine *Engine)StopOneTorrent(hexString string)(stopped bool) {
	singleTorrent, torrentExist := engine.GetOneTorrent(hexString)
	if torrentExist {
		singleTorrentLog, _:= engine.EngineRunningInfo.HashToTorrentLog[singleTorrent.InfoHash()]
		if singleTorrentLog.Status != CompletedStatus {
			singleTorrentLog.Status = StoppedStatus
			//engine.EngineRunningInfo.UpdateTorrentLog()
			singleTorrentLogExtend, extendExist := engine.EngineRunningInfo.TorrentLogExtends[singleTorrent.InfoHash()]
			if extendExist && singleTorrentLogExtend.HasStatusPub && singleTorrentLogExtend.StatusPub != nil {
				singleTorrentLogExtend.HasStatusPub = false
				if !channelClosed(singleTorrentLogExtend.StatusPub.Values) {
					singleTorrentLogExtend.StatusPub.Values <- struct{}{}
					singleTorrentLogExtend.StatusPub.Close()
				}
			}
			singleTorrent.SetMaxEstablishedConns(0)
		}
		stopped = true
	}else{
		stopped = false
	}
	return
}

func (engine *Engine)DelOneTorrent(hexString string)(deleted bool) {
	deleted = false

	//For magnet
	for key, _ := range engine.WebInfo.MagnetTmpInfo {
		if key.HexString() == hexString {
			delete(engine.WebInfo.MagnetTmpInfo, key)
			deleted = true;
			break;
		}
	}


	singleTorrent, torrentExist := engine.GetOneTorrent(hexString)
	if torrentExist {
		deleted = true
		singleTorrent.Drop()
	}
	for index, _:= range engine.EngineRunningInfo.TorrentLogs {
		if engine.EngineRunningInfo.TorrentLogs[index].HashInfoBytes().HexString() == hexString {
			if engine.EngineRunningInfo.TorrentLogs[index].Status == RunningStatus {
				engine.StopOneTorrent(hexString)
			}
			filePath := filepath.Join(engine.EngineRunningInfo.TorrentLogs[index].StoragePath, engine.EngineRunningInfo.TorrentLogs[index].TorrentName)
			log.WithFields(log.Fields{"Path":filePath}).Info("Files have been deleted!")
			engine.EngineRunningInfo.TorrentLogs = append(engine.EngineRunningInfo.TorrentLogs[:index], engine.EngineRunningInfo.TorrentLogs[index+1:]...)
			engine.UpdateInfo()
			delFiles(filePath)
			deleted = true
			break
		}
	}
	if !deleted {
		logger.Error("Not find deleted hash in logs")
	}
	return
}

func delFiles(path string) {
	fmt.Println(path)
	err := os.RemoveAll(path)
	if err != nil {
		logger.WithFields(log.Fields{"Error": err}).Error("unable to delete files")
	}
}


func channelClosed (ch <-chan interface{}) bool {
	select {
	case <-ch:
		return true
	default:
	}

	return false
}