package main

import (
	"github.com/julienschmidt/httprouter"
	"net/http"
	"log"
	"github.com/kisielk/sqlstruct"
	"github.com/skiptirengu/go-mantis-webhook/db"
	"github.com/skiptirengu/go-mantis-webhook/route"
	"github.com/skiptirengu/go-mantis-webhook/config"
	"fmt"
	"strconv"
)

func main() {
	sqlstruct.NameMapper = sqlstruct.ToSnakeCase
	db.Migrate()

	router := httprouter.New()
	router.POST("/webhook/push", route.Middleware.AuthorizeWebhook(route.Webhook.Push))
	router.POST("/app/projects", route.Middleware.App(route.Projects.Add))
	router.POST("/app/aliases", route.Middleware.App(route.Aliases.Add))

	port := fmt.Sprintf(":%s", strconv.Itoa(config.Get().Port))
	log.Printf("Webhook listening on %s", port)
	log.Fatal(http.ListenAndServe(port, router))
}
