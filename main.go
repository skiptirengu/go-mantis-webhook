package main

import (
	"github.com/julienschmidt/httprouter"
	"net/http"
	"log"
)

func main() {
	MigrateDatabase()
	router := httprouter.New()
	router.POST("/push", AuthMiddleware(ReceivePush))
	log.Println("Webhook listening on port 8080")
	log.Fatal(http.ListenAndServe(":8080", router))
}
