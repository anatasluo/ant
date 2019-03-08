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

type TorrentFilePriority struct {
	FilePath		string
	FilePriority  	string
	FileSize		int64
}

type TorrentLocal struct {
	Hash		  		string `storm:"id, unique"`
	InfoBytes	  		[]byte
	DataAdded	  		string
	StoragePath			string
	TempStoragePath		string
	TorrentMoved		bool
	TorrentName			string
	TorrentStatus		string //TODO
	TorrentUploadLimit	bool
	MaxConnections		int
	TorrentType			string
	TorrentFileName		string
	TorrentFile			[]byte
	Label				string
	UploadedBytes		int64
	DownloadedBytes		int64
	TorrentSize			int64
	UploadRatio			string
	TorrentFilePriority	[]TorrentFilePriority
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

func (TorrentQueues *TorrentQueues)UpdateQueues()()  {

}


//REST for Torrent File
func (TorrentDB *TorrentDB)GetAllTorrents()(torrentLocalArray []*TorrentLocal) {
	err := TorrentDB.DB.All(&torrentLocalArray)
	if err != nil {
		logger.WithFields(log.Fields{"Database":TorrentDB.DB, "Detail": err}).Error("Unable To FetchAll")
	}
	return
}

func (TorrentDB *TorrentDB)GetOneTorrent(selectedHash string)(selectedTorrent TorrentLocal) {
	err := TorrentDB.DB.One("Hash", selectedHash, &selectedTorrent)
	if err != nil {
		logger.WithFields(log.Fields{"Database":TorrentDB, "Detail": err}).Error("Error in finding selected torrent")
	}
	return
}

func (TorrentDB *TorrentDB)AddOneTorent(torrentData TorrentLocal)() {
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

func (TorrentDB *TorrentDB)UpdateStorageTick(torrentLocal TorrentLocal)() {
	err := TorrentDB.DB.Update(&torrentLocal)
	if err != nil {
		logger.WithFields(log.Fields{"TorrentInfo":torrentLocal, "Detail": err}).Error("Unable To UpdateOneTorrent")
	}else{
		logger.WithFields(log.Fields{"TorrentInfo":torrentLocal}).Debug("Successfully update one torrent")
	}
}

func (TorrentDB *TorrentDB)GetQueues()(torrentQueues TorrentQueues) {
	err := TorrentDB.DB.One("ID", 5, &torrentQueues)
	if err != nil {
		logger.WithFields(log.Fields{"Database":TorrentDB.DB, "Error":err}).Error("Unable to get queues")
	}
	return
}











