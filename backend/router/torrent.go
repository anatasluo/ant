package router

import (
	"fmt"
	"github.com/julienschmidt/httprouter"
	log "github.com/sirupsen/logrus"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

func addOne(w http.ResponseWriter, r *http.Request, ps httprouter.Params)  {
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

	tmpTorrent, err := runningEngine.AddOneTorrent(filePathAbs)
	if err != nil {
		logger.WithFields(log.Fields{"Error":err}).Error("unable to start a download")
	}
	go func() {
		for {
			<- tmpTorrent.GotInfo()

			per := float64(tmpTorrent.BytesCompleted())  / float64(tmpTorrent.Info().TotalLength())
			fmt.Printf("Current progress : %d / %d --> %.4f \n", tmpTorrent.BytesCompleted(), tmpTorrent.Info().TotalLength(), per)
			time.Sleep(time.Second)
		}
	}()
}

func handleTorrent(router *httprouter.Router)  {
	router.POST("/torrent/addOne", addOne)
}
