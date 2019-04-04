package router

import (
	"fmt"
	"github.com/anatasluo/ant/backend/engine"
	"github.com/julienschmidt/httprouter"
	log "github.com/sirupsen/logrus"
	"io"
	"net/http"
	"os"
	"path/filepath"
)

func addOneTorrentFromFile(w http.ResponseWriter, r *http.Request, ps httprouter.Params)  {

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
	tmpTorrent, err := runningEngine.AddOneTorrentFromFile(filePathAbs)
	
	var isAdded bool
	if err != nil {
		logger.WithFields(log.Fields{"Error":err}).Error("unable to add a torrent")
		isAdded = false
	}else{
		if tmpTorrent != nil {
			runningEngine.GenerateInfoFromTorrent(tmpTorrent)
			runningEngine.StartDownloadTorrent(tmpTorrent.InfoHash().HexString())
			isAdded = true
		}
	}

	WriteResponse(w, JsonFormat{
		"IsAdded":isAdded,
	})

}

func getOneTorrent(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	hexString := r.FormValue("hexString")
	singleTorrent, isExist := runningEngine.GetOneTorrent(hexString)
	if isExist {
		torrentWebInfo := runningEngine.GenerateInfoFromTorrent(singleTorrent)
		WriteResponse(w, torrentWebInfo)
	}else{
		w.WriteHeader(http.StatusNotFound)
	}
}


func appendRunningTorrents(resInfo []engine.TorrentWebInfo)([]engine.TorrentWebInfo) {
	for _, singleTorrent := range runningEngine.TorrentEngine.Torrents() {
		singleTorrentLog, isExist := runningEngine.EngineRunningInfo.HashToTorrentLog[singleTorrent.InfoHash()]
		if isExist && singleTorrentLog.Status != engine.CompletedStatus {
			resInfo = append(resInfo, *runningEngine.GenerateInfoFromTorrent(singleTorrent))
		}
	}
	return resInfo
}

func appendCompletedTorrents(resInfo []engine.TorrentWebInfo)([]engine.TorrentWebInfo) {
	for _, singleTorrentLog := range runningEngine.EngineRunningInfo.TorrentLogs {
		if singleTorrentLog.Status == engine.CompletedStatus {
			resInfo = append(resInfo, *runningEngine.GenerateInfoFromLog(singleTorrentLog))
		}
	}
	return resInfo
}


func getAllTorrents(w http.ResponseWriter, r *http.Request, ps httprouter.Params)  {
	resInfo := []engine.TorrentWebInfo{}
	resInfo = appendRunningTorrents(resInfo)
	resInfo = appendCompletedTorrents(resInfo)
	WriteResponse(w, resInfo)
}

func getCompletedTorrents(w http.ResponseWriter, r *http.Request, ps httprouter.Params)  {
	resInfo := []engine.TorrentWebInfo{}
	resInfo = appendCompletedTorrents(resInfo)
	WriteResponse(w, resInfo)
}



func getAllEngineTorrents(w http.ResponseWriter, r *http.Request, ps httprouter.Params){
	resInfo := []engine.TorrentWebInfo{}
	resInfo = appendRunningTorrents(resInfo)
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
		"IsDownloading":downloaded,
	})
}

func test(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {

}

func handleTorrent(router *httprouter.Router)  {
	router.POST("/torrent/addOneFile", addOneTorrentFromFile)
	router.POST("/torrent/getOne", getOneTorrent)
	router.GET("/torrent/getAllEngineTorrents", getAllEngineTorrents)
	router.GET("/torrent/getAllTorrents", getAllTorrents)
	router.GET("/torrent/getCompletedTorrents", getCompletedTorrents)
	router.POST("/torrent/delOne", delOneTorrent)
	router.POST("/torrent/startDownload", startDownloadTorrent)
	router.POST("/torrent/stopDownload", stopOneTorrent)
	router.GET("/torrent/test", test)
}
