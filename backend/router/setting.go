package router

import (
	"encoding/json"
	"github.com/anatasluo/ant/backend/setting"
	"github.com/julienschmidt/httprouter"
	log "github.com/sirupsen/logrus"
	"net/http"
)

func getSetting(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	WriteResponse(w, clientConfig.GetWebSetting())
}

func applySetting(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	decoder := json.NewDecoder(r.Body)
	isApplied := false
	var newSettings setting.WebSetting
	err := decoder.Decode(&newSettings)
	if err != nil {
		logger.WithFields(log.Fields{"Error": err}).Error("Failed to get new settings")
	}else{
		clientConfig.UpdateConfig(newSettings)
		logger.WithFields(log.Fields{"Settings": newSettings}).Info("Setting update")
		isApplied = true
		runningEngine.Restart()
	}
	WriteResponse(w, JsonFormat{
		"IsApplied":isApplied,
	})
}

func handleSetting(router *httprouter.Router)  {
	router.GET("/settings/config", getSetting)
	router.POST("/settings/apply", applySetting)
}