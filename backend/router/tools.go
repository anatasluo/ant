package router

import (
	"encoding/json"
	"github.com/anacrolix/torrent/metainfo"
	"net/http"
	log "github.com/sirupsen/logrus"
)

type JsonFormat map[string]interface{}

func hexStringToHash(hexString string) (torrentHash metainfo.Hash)  {
	torrentHash = metainfo.Hash{}
	err := torrentHash.FromHexString(hexString)
	if err != nil {
		logger.WithFields(log.Fields{"Error":err}).Error("Unable to get hash from hex string")
	}
	return
}

func WriteResponse(w http.ResponseWriter, jsonStruct interface{}) {
	resInfo, err := json.Marshal(jsonStruct)
	if err != nil {
		logger.WithFields(log.Fields{"Error":err}).Error("unable to format a json")
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(resInfo)
}


