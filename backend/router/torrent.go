package router

import (
	"fmt"
	"github.com/anacrolix/torrent"
	"github.com/anatasluo/ant/backend/engine"
	"github.com/dustin/go-humanize"
	"github.com/julienschmidt/httprouter"
	log "github.com/sirupsen/logrus"
	"io"
	"net/http"
	"os"
	"path/filepath"
)

func addOneTorrent(w http.ResponseWriter, r *http.Request, ps httprouter.Params)  {

	//Get torrent file from form
	err := r.ParseMultipartForm(32 << 20)
	if err != nil {
		logger.WithFields(log.Fields{"Error":err}).Error("Unable to parse form")
		return
	}
	file, handler, err := r.FormFile("oneTorrentFile")

	if err != nil {
		logger.WithFields(log.Fields{"Error":err}).Error("Unable to get file from form")
		return
	}

	defer file.Close()

	filePath := filepath.Join(clientConfig.EngineSetting.Tmpdir, handler.Filename)
	filePathAbs, _ := filepath.Abs(filePath)

	f, err := os.OpenFile(filePathAbs, os.O_WRONLY|os.O_CREATE, 0666)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer f.Close()

	_, err = io.Copy(f, file)
	if err != nil {
		logger.WithFields(log.Fields{"Error":err}).Error("Unable to copy file from form")
		return
	}

	//Start to add to client
	tmpTorrent, err := runningEngine.AddOneTorrent(filePathAbs)

	var jsonFormat JsonFormat
	if err != nil {
		logger.WithFields(log.Fields{"Error":err}).Error("unable to start a download")
		jsonFormat = JsonFormat{
			"Error":"Task has been completed",
		}
	}else{
		jsonFormat = JsonFormat{
			"HexHash":tmpTorrent.InfoHash().HexString(),
		}
	}

	WriteResponse(w, jsonFormat)

}

//TODO Status has been changing during the runing period
func generateInfoFromTorrent(singleTorrent *torrent.Torrent) (torrentWebInfo *engine.TorrentWebInfo) {
	torrentWebInfo, isExist := runningEngine.WebInfo.HashToTorrentWebInfo[singleTorrent.InfoHash()]
	if !isExist {
		torrentLog, _ := runningEngine.EngineRunningInfo.HashToTorrentLog[singleTorrent.InfoHash()]
		torrentWebInfo = &engine.TorrentWebInfo{
			TorrentName	:	singleTorrent.Info().Name,
			TotalLength	:	generateByteSize(singleTorrent.Info().TotalLength()),
			Status		:	engine.StatusIDToName[torrentLog.Status],
			StoragePath	:	torrentLog.StoragePath,
		}
		for _, key := range singleTorrent.Files() {
			torrentWebInfo.Files = append(torrentWebInfo.Files, engine.FileInfo{
				Path	:	key.Path(),
				Priority:	byte(key.Priority()),
				Size	:	generateByteSize(key.Length()),
			})
		}
		runningEngine.WebInfo.HashToTorrentWebInfo[singleTorrent.InfoHash()] = torrentWebInfo
	}else{
		torrentLog, _ := runningEngine.EngineRunningInfo.HashToTorrentLog[singleTorrent.InfoHash()]
		torrentWebInfo.Status = engine.StatusIDToName[torrentLog.Status]
	}
	return
}

func generateByteSize(byteSize int64) string {
	return humanize.Bytes(uint64(byteSize))
}

func getOneTorrent(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	hexString := r.FormValue("hexString")
	singleTorrent, isExist := runningEngine.GetOneTorrent(hexString)
	if isExist {
		torrentWebInfo := generateInfoFromTorrent(singleTorrent)
		WriteResponse(w, torrentWebInfo)
	}else{
		w.WriteHeader(http.StatusNotFound)
	}
}

func getAllTorrent(w http.ResponseWriter, r *http.Request, ps httprouter.Params){
	resInfo := []string{}
	for _, singleTorrentLog := range runningEngine.EngineRunningInfo.TorrentLogs {
		if singleTorrentLog.Status != engine.DeletedStatus && singleTorrentLog.Status != engine.CompletedStatus {
			resInfo = append(resInfo, singleTorrentLog.MetaInfo.HashInfoBytes().HexString())
		}
	}
	WriteResponse(w, resInfo)
}

func delOneTorrent(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	hexString := r.FormValue("hexString")
	deleted := runningEngine.DelOneTorrent(hexString)
	WriteResponse(w, JsonFormat{
		"IsDeleted":deleted,
	})
}

func stopOneTorrent(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	hexString := r.FormValue("hexString")
	stopped := runningEngine.StopOneTorrent(hexString)
	WriteResponse(w, JsonFormat{
		"IsStopped":stopped,
	})
}

func startDownloadTorrent(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	hexString := r.FormValue("hexString")
	downloaded := runningEngine.StartDownloadTorrent(hexString)
	WriteResponse(w, JsonFormat{
		"Downloaded":downloaded,
	})
}

func test(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	torrents := runningEngine.TorrentEngine.Torrents()
	for _, tt := range torrents {
		<- tt.GotInfo()
		fmt.Printf("%+v\n", tt)
	}
	runningEngine.TorrentEngine.WriteStatus(w)
}

func handleTorrent(router *httprouter.Router)  {
	router.POST("/torrent/addOne", addOneTorrent)
	router.POST("/torrent/getOne", getOneTorrent)
	router.GET("/torrent/getAll", getAllTorrent)
	router.POST("/torrent/delOne", delOneTorrent)
	router.POST("/torrent/startDownload", startDownloadTorrent)
	router.POST("/torrent/stopDownload", stopOneTorrent)
	router.GET("/torrent/test", test)
}
