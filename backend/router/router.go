package router

import (
	"github.com/julienschmidt/httprouter"
	"github.com/rs/cors"
	"github.com/urfave/negroni"
)

func InitRouter() *negroni.Negroni {

	router := httprouter.New()
	handleAuth(router)

	// Use global middleware
	n := negroni.New()

	//Enable cors
	c := cors.Default()

	n.Use(c)
	n.Use(negroni.NewLogger())
	n.UseHandler(router)

	return n
}