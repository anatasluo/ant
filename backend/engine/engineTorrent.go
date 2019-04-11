package engine

import (
	"errors"
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
			engine.SaveInfo()
		}
	}
	return tmpTorrent, err
}

//Check if duplicated torrent
func (engine *Engine)checkOneHash(infoHash metainfo.Hash) (tmpTorrent *torrent.Torrent, needMoreOperation bool){
	torrentLog, isExist := engine.EngineRunningInfo.HashToTorrentLog[infoHash]
	if isExist && torrentLog.Status != CompletedStatus {
		logger.Info("Task has been created")
		tmpTorrent, _ = engine.TorrentEngine.Torrent(infoHash)
		needMoreOperation = false
	}else if isExist && torrentLog.Status == CompletedStatus {
		logger.Info("Task has been completed")
		tmpTorrent = nil
		needMoreOperation = false
	}else{
		logger.Info("Create a new task")
		needMoreOperation = true
	}
	return
}

func (engine *Engine)AddOneTorrentFromMagnet (linkAddress string)(tmpTorrent *torrent.Torrent, err error) {
	if strings.HasPrefix(linkAddress, "magnet:") || strings.HasPrefix(linkAddress, "infohash:") {
		var infoHash metainfo.Hash
		if strings.HasPrefix(linkAddress, "magnet:") {
			var torrentMetaInfo *torrent.TorrentSpec
			torrentMetaInfo, err = torrent.TorrentSpecFromMagnetURI(linkAddress)
			if err != nil {
				logger.WithFields(log.Fields{"Error":err}).Error("unable to resolve magnet")
				return
			}else{
				infoHash = torrentMetaInfo.InfoHash
			}
		} else {
			infoHash = metainfo.NewHashFromHex(strings.TrimPrefix(linkAddress, "infohash:"))
		}
		var needMoreOperation bool
		tmpTorrent, needMoreOperation = engine.checkOneHash(infoHash)

		if needMoreOperation {
			engine.EngineRunningInfo.AddOneTorrentFromMagnet(infoHash)
			extendLog, _ := engine.EngineRunningInfo.TorrentLogExtends[infoHash]
			go func() {
				engine.EngineRunningInfo.MagnetNum ++
				tmpTorrent, err = engine.TorrentEngine.AddMagnet(linkAddress)
				select {
					case <- tmpTorrent.GotInfo():
						if err != nil {
							logger.WithFields(log.Fields{"Error":err, "Torrent": tmpTorrent}).Error("Unable to resolve magnet")
						}else{
							logger.Debug("Add torrent from magnet")
							engine.EngineRunningInfo.UpdateMagnetInfo(tmpTorrent)
							engine.SaveInfo()
							engine.GenerateInfoFromTorrent(tmpTorrent)
							engine.StartDownloadTorrent(tmpTorrent.InfoHash().HexString())
							engine.EngineRunningInfo.EngineCMD <- RefreshInfo
							logger.Debug("It should refresh")
						}
					case <- extendLog.MagnetAnalyseChan:
						tmpTorrent.Drop()
						extendLog.MagnetDelChan <- true
						logger.Debug("One magnet has been deleted")
				}
				engine.EngineRunningInfo.MagnetNum --
			}()
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
		tmpTorrent = nil
		isExist = false
	}else{
		tmpTorrent, isExist = engine.TorrentEngine.Torrent(torrentHash)

		//any operation on magnet is forbidden
		if isExist {
			torrentLog, _ := engine.EngineRunningInfo.HashToTorrentLog[tmpTorrent.InfoHash()]
			if torrentLog.Status == AnalysingStatus {
				tmpTorrent = nil
				isExist = false
			}
		}
	}
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
			engine.SaveInfo()
			//check if extend exist
			_, extendIsExist := engine.EngineRunningInfo.TorrentLogExtends[singleTorrent.InfoHash()];
			if !extendIsExist {
				engine.EngineRunningInfo.TorrentLogExtends[singleTorrent.InfoHash()] = &TorrentLogExtend{
					StatusPub:singleTorrent.SubscribePieceStateChanges(),
					HasStatusPub:true,
					HasMagnetChan:false,
				}
			}else if extendIsExist && !engine.EngineRunningInfo.TorrentLogExtends[singleTorrent.InfoHash()].HasStatusPub{
				logger.Debug("it has extend but no status pub")
				engine.EngineRunningInfo.TorrentLogExtends[singleTorrent.InfoHash()].HasStatusPub = true
				engine.EngineRunningInfo.TorrentLogExtends[singleTorrent.InfoHash()].StatusPub = singleTorrent.SubscribePieceStateChanges()
			}
			logger.Debug("Create extend info for log")
			//Some download setting for task
			//logger.Debug(clientConfig.DefaultTrackers)
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
		engine.SaveInfo()
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

// TODO: Find error of out range of index, not find reason now
// Delete on torrent will operate logs directly, rather than get from getOne
func (engine *Engine)DelOneTorrent(hexString string)(deleted bool) {
	deleted = false

	for index := range engine.EngineRunningInfo.TorrentLogs {
		if engine.EngineRunningInfo.TorrentLogs[index].Status != AnalysingStatus && engine.EngineRunningInfo.TorrentLogs[index].HashInfoBytes().HexString() == hexString {
			if engine.EngineRunningInfo.TorrentLogs[index].Status == RunningStatus {
				engine.StopOneTorrent(hexString)
			}
			singleTorrent, torrentExist := engine.TorrentEngine.Torrent(engine.EngineRunningInfo.TorrentLogs[index].HashInfoBytes())
			if torrentExist {
				singleTorrent.Drop()
			}
			filePath := filepath.Join(engine.EngineRunningInfo.TorrentLogs[index].StoragePath, engine.EngineRunningInfo.TorrentLogs[index].TorrentName)
			logger.WithFields(log.Fields{"Path":filePath}).Info("Files have been deleted!")
			engine.EngineRunningInfo.TorrentLogs = append(engine.EngineRunningInfo.TorrentLogs[:index], engine.EngineRunningInfo.TorrentLogs[index+1:]...)
			engine.UpdateInfo()
			delFiles(filePath)
			deleted = true
			return
		}else if engine.EngineRunningInfo.TorrentLogs[index].Status == AnalysingStatus && engine.EngineRunningInfo.TorrentLogs[index].TorrentName == hexString {

			//Magnet hash is stored in torrentName
			torrentHash := metainfo.Hash{}
			_ = torrentHash.FromHexString(engine.EngineRunningInfo.TorrentLogs[index].TorrentName)
			extendLog, _ := engine.EngineRunningInfo.TorrentLogExtends[torrentHash]
			extendLog.MagnetAnalyseChan <- true
			select {
				case <- extendLog.MagnetDelChan:
					engine.EngineRunningInfo.TorrentLogs = append(engine.EngineRunningInfo.TorrentLogs[:index], engine.EngineRunningInfo.TorrentLogs[index+1:]...)
					engine.UpdateInfo()
					deleted = true
					logger.Debug("Delete Magnet Done")
					return
			}
		}
	}
	engine.SaveInfo()
	return
}

func delFiles(path string) {
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