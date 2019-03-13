package router

import (
	"github.com/gorilla/websocket"
	"github.com/julienschmidt/httprouter"
	"net/http"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func torrentProgress (w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	hexString := ps.ByName("hexString")
	singleTorrent, isExist := runningEngine.GetOneTorrent(hexString)
	if !isExist {
		w.WriteHeader(http.StatusNotFound)
	}else{
		singleTorrentLogExtend, extendExist := runningEngine.EngineRunningInfo.TorrentLogExtends[singleTorrent.InfoHash()]
		if !extendExist {
			w.WriteHeader(http.StatusNotFound)
		} else {
			singleTorrentLogExtend.WebNeed = true
			conn, err := upgrader.Upgrade(w, r, nil)
			if err != nil {
				logger.Error("Unable to init websocket", err)
				return
			}
			defer conn.Close()
			for {
				torrentProgressInfo := <- singleTorrentLogExtend.ProgressInfo
				err = conn.WriteJSON(torrentProgressInfo)
				if err != nil {
					logger.Error("Unable to write Message", err)
					break
				}
			}
			singleTorrentLogExtend.WebNeed = false
		}

	}

}

func handleWS (router *httprouter.Router)  {
	router.GET("/ws/:hexString", torrentProgress)
}

