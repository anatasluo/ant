package router

import (
	"github.com/julienschmidt/httprouter"
	"net/http"
)

func addOneMagent(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {

}

func handleMagent(router *httprouter.Router)  {
	router.POST("/magent/addOne", addOneMagent)
}