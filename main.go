package main

import (
	"github.com/julienschmidt/httprouter"
	"net/http"
	"log"
	"github.com/kisielk/sqlstruct"
	"github.com/skiptirengu/go-mantis-webhook/route"
	"github.com/skiptirengu/go-mantis-webhook/config"
	"fmt"
	"strconv"
	"github.com/skiptirengu/go-mantis-webhook/db"
)

func main() {
	var (
		conf       = config.Get()
		router     = httprouter.New()
		middleware = route.NewMiddleware(conf)
		database   = db.Get()
	)

	sqlstruct.NameMapper = sqlstruct.ToSnakeCase
	database.Migrate()

	router.POST("/webhook/push", middleware.AuthorizeWebhook(route.Webhook(conf, database).Push))
	router.POST("/app/projects", middleware.App(route.Projects(database).Add))
	router.POST("/app/aliases", middleware.App(route.Aliases(database).Add))

	port := fmt.Sprintf(":%s", strconv.Itoa(conf.Port))
	log.Printf("Webhook listening on %s", port)
	log.Fatal(http.ListenAndServe(port, router))
}
