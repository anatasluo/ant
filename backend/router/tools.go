package router

import (
	"encoding/json"
	"net/http"
	log "github.com/sirupsen/logrus"
)

type JsonFormat map[string]interface{}

func WriteResponse(w http.ResponseWriter, jsonStruct interface{}) {
	resInfo, err := json.Marshal(jsonStruct)
	if err != nil {
		logger.WithFields(log.Fields{"Error":err}).Error("unable to format a json")
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(resInfo)
}


