package main

import (
	"context"
	"database/sql"
	"github.com/aalug/go-gin-job-search/internal/api"
	"github.com/aalug/go-gin-job-search/internal/config"
	"github.com/aalug/go-gin-job-search/internal/db/sqlc"
	"github.com/aalug/go-gin-job-search/internal/esearch"
	"log"
)

func main() {
	// === config, env file ===
	cfg, err := config.LoadConfig(".")
	if err != nil {
		log.Fatal("cannot load env file: ", err)
	}

	// === database ===
	conn, err := sql.Open(cfg.DBDriver, cfg.DBSource)
	if err != nil {
		log.Fatal("cannot connect to the db: ", err)
	}

	store := db.NewStore(conn)

	// === Elasticsearch ===
	ctx := context.Background()
	ctx, err = esearch.LoadJobsFromDB(ctx, store)
	if err != nil {
		log.Fatal("cannot load jobs from db: ", err)
	}
	newClient, err := esearch.ConnectWithElasticsearch(cfg.ElasticSearchAddress)
	if err != nil {
		log.Fatal("cannot connect to the elasticsearch: ", err)
	}

	client := esearch.NewClient(newClient)
	err = client.IndexJobsAsDocuments(ctx)
	if err != nil {
		log.Fatal("cannot index jobs as documents: ", err)
	}

	// === HTTP server ===
	server, err := api.NewServer(cfg, store, client)
	if err != nil {
		log.Fatal("cannot create server: ", err)
	}

	err = server.Start(cfg.ServerAddress)
	if err != nil {
		log.Fatal("cannot start the server:", err)
	}
}
