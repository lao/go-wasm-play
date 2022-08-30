package main

import (
	"fmt"
	"net/http"

	"github.com/NYTimes/gziphandler"
)

var serverPort = ":9090"

func main() {
	fmt.Println("Starting server at: http://localhost" + serverPort)
	err := http.ListenAndServe(serverPort, gziphandler.GzipHandler(http.FileServer(http.Dir("../../assets"))))
	if err != nil {
		fmt.Println("Failed to start server", err)
		return
	}
}
