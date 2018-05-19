package mantis

import (
	"log"
	"github.com/skiptirengu/go-mantis-webhook/db"
	"github.com/skiptirengu/go-mantis-webhook/config"
	"github.com/pkg/errors"
	"fmt"
)

func getHost(conf *config.Configuration) (host string) {
	if host = conf.Mantis.Host; host == "" {
		log.Fatal("Mantis host is empty")
	}
	return
}

func SyncProjectUsers(c *config.Configuration, db db.Database, projectName string) (error) {
	service := NewSoapService(c)
	projectId, err := service.ProjectGetIdFromName(projectName)

	if err != nil {
		return err
	} else if projectId == 0 {
		return errors.New(fmt.Sprintf("project %s not found", projectName))
	}

	accounts, err := service.ProjectGetUsers(projectId)
	if err != nil {
		return err
	}

	for _, account := range accounts {
		if _, err := db.Users().CreateOrUpdate(account.Id, account.Name, account.Email); err != nil {
			return err
		}
	}

	return nil
}
