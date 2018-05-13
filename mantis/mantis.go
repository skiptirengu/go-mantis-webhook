package mantis

import (
	"github.com/skiptirengu/go-mantis-webhook/config"
	"log"
	"github.com/skiptirengu/go-mantis-webhook/db"
)

func getHost() (host string) {
	if host = config.Get().Mantis.Host; host == "" {
		log.Fatal("Mantis host is empty")
	}
	return
}

func SyncProjectUsers(projectName string) {
	projectId, err := Soap.ProjectGetIdFromName(projectName)

	if err != nil {
		log.Println(err)
		return
	} else if projectId == 0 {
		log.Printf("Project %s not found", projectName)
		return
	}

	accounts, err := Soap.ProjectGetUsers(projectId)
	if err != nil {
		log.Println(err)
		return
	}

	for _, account := range accounts {
		if _, err := db.Users.CreateOrUpdate(account.Id, account.Name, account.Email); err != nil {
			log.Println(err)
		}
	}
}
