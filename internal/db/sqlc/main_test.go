package db

import (
	"database/sql"
	"github.com/aalug/job-finder-go/internal/config"
	"log"
	"os"
	"testing"

	_ "github.com/lib/pq"
)

var testQueries *Queries
var testStore Store
var testDB *sql.DB

func TestMain(m *testing.M) {
	var err error

	cfg, err := config.LoadConfig("../../../.")
	if err != nil {
		log.Fatal("cannot load env file: ", err)
	}

	testDB, err = sql.Open(cfg.DBDriver, cfg.DBSource)
	if err != nil {
		log.Fatal("cannot connect to db: ", err)
	}

	testStore = NewStore(testDB)
	testQueries = New(testDB)

	os.Exit(m.Run())
}
