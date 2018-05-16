package db

import (
	_ "github.com/lib/pq"

	"sync"
	"fmt"
	"database/sql"
	"log"
	"github.com/skiptirengu/go-mantis-webhook/config"
	"sort"
	"strings"
	"io/ioutil"
	"path"
	"github.com/kisielk/sqlstruct"
	"os"
	"time"
)

var (
	connectionMu = sync.Mutex{}
	con          *sql.DB
)

type Database interface {
	Aliases() (*aliases)
	Issues() (*issues)
	Projects() (*projects)
	Users() (*users)
	GetDB() (*sql.DB)
	Migrate()
}

type Migration struct {
	Version   string
	Timestamp time.Time
}

type Connection struct {
	db   *sql.DB
	conf *config.Configuration
}

func (c *Connection) GetDB() (*sql.DB) {
	return c.db
}

func (c *Connection) Aliases() (*aliases) {
	return &aliases{c.db}
}

func (c *Connection) Issues() (*issues) {
	return &issues{c.db}
}

func (c *Connection) Projects() (*projects) {
	return &projects{c.db}
}

func (c *Connection) Users() (*users) {
	return &users{c.db}
}

func Get() (Database) {
	c := config.Get()
	con := Connection{connectDatabase(c), c}
	return &con
}

func (c *Connection) getAppliedMigrations() (map[string]*Migration) {
	var (
		db     = c.GetDB()
		rows   = make(map[string]*Migration)
		err    error
		dbRows *sql.Rows
	)

	c.createMigrationTable()

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

func (c *Connection) createMigrationTable() {
	_, err := c.GetDB().Exec(`
		create table if not exists migrations (
  			version   varchar not null,
  			timestamp timestamp without time zone default (now() at time zone 'utc')
		);
	`)

	if err != nil {
		log.Fatal(err)
	}
}

func (c *Connection) getMigrationsToApply() ([]os.FileInfo) {
	files, err := ioutil.ReadDir("migrations")

	if err != nil {
		log.Fatal(err)
	}

	sort.Slice(files, func(i, j int) bool {
		return strings.Compare(files[i].Name(), files[j].Name()) == -1
	})

	return files
}

func (c *Connection) Migrate() {
	var (
		db      = c.GetDB()
		files   = c.getMigrationsToApply()
		applied = c.getAppliedMigrations()
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
		}

		if _, err := t.Exec("insert into migrations (version) values ($1)", file.Name()); err != nil {
			t.Rollback()
			log.Fatal(err)
		}

		log.Println(fmt.Sprintf("Applied migration %s", file.Name()))
		t.Commit()
	}
}

func connectDatabase(c *config.Configuration) (*sql.DB) {
	connectionMu.Lock()
	defer connectionMu.Unlock()

	var (
		conf    = c.Database
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

func ScanCol(r *sql.Rows, v interface{}) (error) {
	r.Next()
	return r.Scan(v)
}

func NullableStr(val string) (sql.NullString) {
	if len(val) == 0 {
		return sql.NullString{}
	}
	return sql.NullString{
		Valid:  true,
		String: val,
	}
}
