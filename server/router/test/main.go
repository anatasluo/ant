package router

import (
	"fmt"
	"net/http"
)


func sayhelloName(w http.ResponseWriter, r *http.Request) {
	//config.SetupResponse(&w, r)
	fmt.Println(r.Method)

	fmt.Fprintf(w, "Hello astaxie!")
}

func Test() {
	http.HandleFunc("/test", sayhelloName)
}
