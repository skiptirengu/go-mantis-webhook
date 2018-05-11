package mantis

import (
	"github.com/skiptirengu/go-mantis-webhook/config"
	"log"
)

func getHost() (host string) {
	if host = config.Get().Mantis.Host; host == "" {
		log.Fatal("Mantis host is empty")
	}
	return
}
