package mantis

import (
	"log"
	"github.com/skiptirengu/go-mantis-webhook/db"
	"github.com/skiptirengu/go-mantis-webhook/config"
)

func getHost(conf *config.Configuration) (host string) {
	if host = conf.Mantis.Host; host == "" {
		log.Fatal("Mantis host is empty")
	}
	return
}

func SyncProjectUsers(c *config.Configuration, db db.Database, projectName string) {
	service := NewSoapService(c)
	projectId, err := service.ProjectGetIdFromName(projectName)

	if err != nil {
		log.Println(err)
		return
	} else if projectId == 0 {
		log.Printf("Project %s not found", projectName)
		return
	}

	accounts, err := service.ProjectGetUsers(projectId)
	if err != nil {
		log.Println(err)
		return
	}

	for _, account := range accounts {
		if _, err := db.Users().CreateOrUpdate(account.Id, account.Name, account.Email); err != nil {
			log.Println(err)
		}
	}
}
