package main

import (
	"fmt"
	"github.com/anatasluo/ant/server/router"
	"github.com/zserge/webview"
	"log"
	"net"
	"net/http"
	"strings"
	"sync"
)

type server_config struct {
	ln *net.Listener
}

type client_config struct {
	path string
}

var Server_config server_config

var Client_config client_config

var wg sync.WaitGroup


func configInit()  {
	Client_config.path = "/home/anatas/Desktop/Git/ant/webview/dist/webview"
}

func runLocalHTTP() {
	// let os decides the port
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		log.Fatal(err)
	}

	Server_config.ln = &ln
	go func() {
		defer  ln.Close()
		// Init server router
		router.Run()

		// For webpage
		fs := http.FileServer(http.Dir(Client_config.path))
		http.Handle("/", fs)


		log.Fatal(http.Serve(ln, nil))

	}()

}


func runWebview() {

	ln := *Server_config.ln

	w := webview.New(webview.Settings{
		Title: "ANT Downloader",
		URL:   "http://" + ln.Addr().String() + "/index.html",
		Width: 950,
		Height: 800,
		Resizable:true,
	})

	defer w.Exit()
	fmt.Println(ln.Addr().String())
	// transform params here for webview
	port := "const port = " + strings.Split((*Server_config.ln).Addr().String(), ":")[1]
	w.Eval(port)

	w.Run()

}


func main() {
	configInit()
	runLocalHTTP()
	runWebview()
}

