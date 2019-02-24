package router

import (
	"fmt"
	"github.com/anatasluo/ant/backend/setting"
	"github.com/julienschmidt/httprouter"
	"net/http"
)

var (
	clientConfig = setting.GetClientSetting()
)

func handleLogin(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	fmt.Printf("%+v\n", ps)
	//userstate.Login(w, "bob")
}

func handleAuth(router *httprouter.Router) {


	router.POST("/login", handleLogin)
}
