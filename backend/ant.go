package main

import (
	"github.com/anatasluo/ant/backend/engine"
	"github.com/anatasluo/ant/backend/router"
	"github.com/anatasluo/ant/backend/setting"
	log "github.com/sirupsen/logrus"
	"github.com/urfave/negroni"
	"net/http"
	"os"
	"os/signal"
	"runtime"
)

var (
	clientConfig 		= setting.GetClientSetting()
	logger 				= clientConfig.LoggerSetting.Logger
	torrentEngine 		= engine.GetEngine()
	nRouter				  *negroni.Negroni
)

func runAPP() {
	go func() {
		// Init server router
		nRouter = router.InitRouter()
		err := http.ListenAndServe(clientConfig.ConnectSetting.Addr, nRouter)
		if err != nil {
			logger.WithFields(log.Fields{"Error":err}).Fatal("Failed to created http service")
		}

	}()
}

func cleanUp()  {
	go func() {
		c := make(chan os.Signal, 1)
		signal.Notify(c, os.Interrupt)
		<-c
		log.Info("The progame will stop!")
		torrentEngine.Cleanup()
		os.Exit(0)
	}()
}

func test()  {

}
func main() {
	runAPP()
	cleanUp()
	test()
	runtime.Goexit()
}

