package main

import (
	"github.com/julienschmidt/httprouter"
	"net/http"
	"log"
)

func main() {
	router := httprouter.New()
	router.POST("/push", AuthMiddleware(ReceivePush))
	log.Println("Starting webhook on port :8080")
	log.Fatal(http.ListenAndServe(":8080", router))
}
