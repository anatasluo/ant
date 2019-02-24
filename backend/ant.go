package main

import (
	"github.com/anatasluo/ant/backend/router"
	"github.com/anatasluo/ant/backend/setting"
	log "github.com/sirupsen/logrus"
	"net/http"
	"sync"
)

var (
	Logger = log.New()
	ClientConfig = setting.GetClientSetting()

	Wg sync.WaitGroup
)

func runLocalHTTP() {
	Wg.Add(1)

	go func() {
		// Init server router
		n := router.InitRouter()
		log.Fatal(http.ListenAndServe(ClientConfig.ConnectSetting.Addr, n))
	}()

	Wg.Wait()
}


func main() {

	runLocalHTTP()

}

