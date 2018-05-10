package main

import (
	_ "github.com/lib/pq"

	"sync"
	"fmt"
	"database/sql"
	"log"
	"io/ioutil"
	"sort"
	"strings"
	"os"
	"path"
	"time"
	"github.com/kisielk/sqlstruct"
)

var (
	databaseMutex = sync.Mutex{}
	con           *sql.DB
)

func GetDB() (*sql.DB) {
	databaseMutex.Lock()
	defer databaseMutex.Unlock()

	var (
		conf    = GetConfig().Database
		connStr = fmt.Sprintf("postgres://%s:%s@%s/%s?sslmode=disable", conf.User, conf.Password, conf.Host, conf.DatabaseName)
		err     error
	)

	if con != nil {
		return con
	}

	if con, err = sql.Open("postgres", connStr); err != nil {
		log.Fatal(err)
	}

	return con
}

func getAppliedMigrations() (map[string]*Migration) {
	var (
		db     = GetDB()
		rows   = make(map[string]*Migration)
		err    error
		dbRows *sql.Rows
	)

	createMigrationTable()

	if dbRows, err = db.Query("select version, timestamp from migrations"); err != nil {
		log.Fatal(err)
	}

	for dbRows.Next() {
		migration := Migration{}
		sqlstruct.Scan(&migration, dbRows)
		rows[migration.Version] = &migration
	}

	return rows
}

func createMigrationTable() {
	_, err := GetDB().Exec(`
		create table if not exists migrations (
  			version   varchar not null,
  			timestamp timestamp without time zone default (now() at time zone 'utc')
		);
	`)

	if err != nil {
		log.Fatal(err)
	}
}

func getMigrationsToApply() ([]os.FileInfo) {
	files, err := ioutil.ReadDir("migrations")

	if err != nil {
		log.Fatal(err)
	}

	sort.Slice(files, func(i, j int) bool {
		return strings.Compare(files[i].Name(), files[j].Name()) == -1
	})

	return files
}

func MigrateDatabase() {
	var (
		db      = GetDB()
		files   = getMigrationsToApply()
		applied = getAppliedMigrations()
	)

	for _, file := range files {
		if _, b := applied[file.Name()]; b {
			continue
		}

		bytes, err := ioutil.ReadFile(path.Join("migrations", file.Name()))
		if err != nil {
			log.Fatal(err)
		}

		t, err := db.Begin()
		if err != nil {
			log.Fatal(err)
		}

		for _, script := range strings.Split(string(bytes), ";") {
			if strings.TrimSpace(script) == "" {
				continue
			}
			if _, err := t.Exec(script); err != nil {
				t.Rollback()
				log.Fatal(err)
			}
			if _, err := t.Exec("insert into migrations (version) values ($1)", file.Name()); err != nil {
				t.Rollback()
				log.Fatal(err)
			}
			log.Println(fmt.Sprintf("Applied migration %s", file.Name()))
		}

		t.Commit()
	}
}

type Migration struct {
	Version   string
	Timestamp time.Time
}
