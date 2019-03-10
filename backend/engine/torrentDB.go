package engine

import (
	"github.com/asdine/storm"
	log "github.com/sirupsen/logrus"
	"os"
	"path/filepath"
)

type TorrentDB struct {
	DB 		*storm.DB
	Path	string
}

func GetTorrentDB(dbPath string) *TorrentDB {
	db, err := storm.Open(dbPath)
	var torrentDB	TorrentDB
	torrentDB.DB = db
	torrentDB.Path = dbPath
	if err != nil {
		log.WithFields(log.Fields{"Error": err, "Path": torrentDB.Path}).Fatal("Failed to create database for engine")
	}
	return &torrentDB
}

func (TorrentDB *TorrentDB)Cleanup()() {
	if TorrentDB.DB != nil {
		err := TorrentDB.DB.Close()
		if err != nil {
			logger.WithFields(log.Fields{"Detail":err}).Error("Failed to closed database")
		}
	}
}

func (TorrentQueues *TorrentLogsAndID)UpdateQueues()()  {

}


//REST for Torrent File
func (TorrentDB *TorrentDB)GetAllTorrents()(TorrentLogArray []*TorrentLog) {
	err := TorrentDB.DB.All(&TorrentLogArray)
	if err != nil {
		logger.WithFields(log.Fields{"Database":TorrentDB.DB, "Detail": err}).Error("Unable To FetchAll")
	}
	return
}

func (TorrentDB *TorrentDB)GetOneTorrent(selectedHash string)(selectedTorrent TorrentLog) {
	err := TorrentDB.DB.One("Hash", selectedHash, &selectedTorrent)
	if err != nil {
		logger.WithFields(log.Fields{"Database":TorrentDB, "Detail": err}).Error("Error in finding selected torrent")
	}
	return
}

func (TorrentDB *TorrentDB)AddOneTorent(torrentData TorrentLog)() {
	logger.WithFields(log.Fields{"path" : torrentData.StoragePath, "name": torrentData.TorrentName}).Info("Add One Torrent successful")
	err := TorrentDB.DB.Save(&torrentData)
	if err != nil {
		logger.WithFields(log.Fields{"Database":TorrentDB.DB, "Detail": err}).Error("Unable To AddOneTorrent")
	}
}

func (TorrentDB *TorrentDB)DelOneTorrent(SelectedHash string)(selectedTorrentName string) {
	selectedTorrent := TorrentDB.GetOneTorrent(SelectedHash)
	selectedTorrentName = selectedTorrent.TorrentName
	err := TorrentDB.DB.DeleteStruct(&selectedTorrent)
	if err != nil {
		logger.WithFields(log.Fields{"Database":TorrentDB, "Detail": err}).Error("Error in deleting selected torrent")
	}
	return
}

func (TorrentDB *TorrentDB)DelOneTorrentAndFile(selectedHash string, downloadPath string)(selectedTorrentName string) {
	selectedTorrentName = TorrentDB.DelOneTorrent(selectedHash)
	torrentFilePath := filepath.Join(downloadPath, selectedTorrentName)
	err := os.RemoveAll(torrentFilePath)
	if err != nil {
		logger.WithFields(log.Fields{"FilePath":torrentFilePath, "Detail":err}).Error("can not delete selected file")
	}
	return
}

func (TorrentDB *TorrentDB)UpdateStorageTick(TorrentLog TorrentLog)() {
	err := TorrentDB.DB.Update(&TorrentLog)
	if err != nil {
		logger.WithFields(log.Fields{"TorrentInfo":TorrentLog, "Detail": err}).Error("Unable To UpdateOneTorrent")
	}else{
		logger.WithFields(log.Fields{"TorrentInfo":TorrentLog}).Debug("Successfully update one torrent")
	}
}

func (TorrentDB *TorrentDB)GetQueues()(torrentLogs TorrentLogsAndID) {
	err := TorrentDB.DB.One("ID", TorrentLogsID, &torrentLogs)
	if err != nil {
		logger.WithFields(log.Fields{"Database":TorrentDB.DB, "Error":err}).Error("Unable to get queues")
	}
	return
}











