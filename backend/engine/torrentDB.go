package engine

import (
	"github.com/asdine/storm"
	log "github.com/sirupsen/logrus"
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

func (TorrentDB *TorrentDB)GetLogs(torrentLogs *TorrentLogsAndID)() {
	torrentLogs.ID = TorrentLogsID;
	err := TorrentDB.DB.One("ID", TorrentLogsID, torrentLogs)
	if err != nil {
		logger.WithFields(log.Fields{"Error":err}).Info("Init running queue now")
	}
	return
}









