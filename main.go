package main

import (
	"github.com/julienschmidt/httprouter"
	"net/http"
	"log"
	"github.com/kisielk/sqlstruct"
	"github.com/skiptirengu/go-mantis-webhook/db"
	"github.com/skiptirengu/go-mantis-webhook/route"
)

func main() {
	sqlstruct.NameMapper = sqlstruct.ToSnakeCase
	db.Migrate()

	router := httprouter.New()
	router.POST("/push", route.Middleware.AuthorizeWebhook(route.Webhook.Push))
	router.POST("/projects", route.Middleware.App(route.Projects.Add))
	router.POST("/aliases", route.Middleware.App(route.Aliases.Add))

	log.Println("Webhook listening on port 8080")
	log.Fatal(http.ListenAndServe(":8080", router))
}
