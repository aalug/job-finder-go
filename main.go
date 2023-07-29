package main

import (
	"database/sql"
	"github.com/aalug/go-gin-job-search/api"
	db "github.com/aalug/go-gin-job-search/db/sqlc"
	"github.com/aalug/go-gin-job-search/esearch"
	"github.com/aalug/go-gin-job-search/utils"
	"log"
)

func main() {
	// === config, env file ===
	config, err := utils.LoadConfig(".")
	if err != nil {
		log.Fatal("cannot load env file: ", err)
	}

	// === database ===
	conn, err := sql.Open(config.DBDriver, config.DBSource)
	if err != nil {
		log.Fatal("cannot connect to the db: ", err)
	}

	store := db.NewStore(conn)

	// === Elasticsearch ===
	//ctx := context.Background()
	// TODO: for now, all jobs are indexed every time the server starts
	// TODO: later on, we will index only new or updated jobs
	//ctx = esearch.LoadJobsFromDB(ctx, store)
	//esearch.IndexJobsAsDocuments(ctx)
	// === elasticsearch ===
	newClient, err := esearch.ConnectWithElasticsearch(config.ElasticSearchAddress)
	if err != nil {
		log.Fatal("cannot connect to the elasticsearch: ", err)
	}

	client := esearch.NewClient(newClient)

	// === HTTP server ===
	// @BasePath /api/v1
	// @contact.name aalug
	// @contact.url https://github.com/aalug
	// @contact.email a.a.gulczynski@gmail.com
	server, err := api.NewServer(config, store, client)
	if err != nil {
		log.Fatal("cannot create server: ", err)
	}

	err = server.Start(config.ServerAddress)
	if err != nil {
		log.Fatal("cannot start the server:", err)
	}
}
