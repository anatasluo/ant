package router

import (
	"github.com/julienschmidt/httprouter"
	log "github.com/sirupsen/logrus"
	"net/http"
)

//Add magnet will let to serious problems, a better way is to get torrent file via magnet and then use addTorrent
func addOneMagnet(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {

	linkAddress := r.FormValue("linkAddress")

	logger.Debug(linkAddress)
	
	_, err := runningEngine.AddOneTorrentFromMagnet(linkAddress)

	var isAdded bool
	if err != nil {
		logger.WithFields(log.Fields{"Error":err}).Error("unable to add a magnet")
		isAdded = false
	}else{
		isAdded = true
	}

	WriteResponse(w, JsonFormat{
		"IsAdded":isAdded,
	})
}

func handleMagnet(router *httprouter.Router)  {
	router.POST("/magnet/addOneMagent", addOneMagnet)
}