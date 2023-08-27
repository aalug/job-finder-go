package main

import (
	"context"
	"database/sql"
	"flag"
	"github.com/aalug/go-gin-job-search/internal/api"
	"github.com/aalug/go-gin-job-search/internal/config"
	"github.com/aalug/go-gin-job-search/internal/db/sqlc"
	"github.com/aalug/go-gin-job-search/internal/esearch"
	"github.com/aalug/go-gin-job-search/internal/mail"
	"github.com/aalug/go-gin-job-search/internal/worker"
	"github.com/hibiken/asynq"
	zerolog "github.com/rs/zerolog/log"
)

func main() {
	// === config, env file ===
	cfg, err := config.LoadConfig(".")
	if err != nil {
		zerolog.Fatal().Err(err).Msg("cannot load env file")
	}

	// === database ===
	conn, err := sql.Open(cfg.DBDriver, cfg.DBSource)
	if err != nil {
		zerolog.Fatal().Err(err).Msg("cannot connect to the db")
	}

	store := db.NewStore(conn)

	// === redis ===
	redisOpt := asynq.RedisClientOpt{
		Addr: cfg.RedisAddress,
	}

	taskDistributor := worker.NewRedisTaskDistributor(redisOpt)

	// === loading test data ===
	loadDataFlag := flag.Bool("load_test_data", false, "If set, the application will load test data into db")
	flag.Parse()

	if *loadDataFlag != false {
		// load the test data into the db
		store.LoadTestData(context.Background())
	}

	// === Elasticsearch ===
	ctx := context.Background()
	ctx, err = esearch.LoadJobsFromDB(ctx, store)
	if err != nil {
		zerolog.Fatal().Err(err).Msg("cannot load jobs from db")
	}
	newClient, err := esearch.ConnectWithElasticsearch(cfg.ElasticSearchAddress)
	if err != nil {
		zerolog.Fatal().Err(err).Msg("cannot connect to the elasticsearch")
	}

	client := esearch.NewClient(newClient)
	err = client.IndexJobsAsDocuments(ctx)
	if err != nil {
		zerolog.Fatal().Err(err).Msg("cannot index jobs as documents")
	}

	// === task processor ===
	go runTaskProcessor(cfg, redisOpt, store)
	// === HTTP server ===
	runHTTPServer(cfg, store, client, taskDistributor)
}

func runHTTPServer(cfg config.Config, store db.Store, client esearch.ESearchClient, taskDistributor worker.TaskDistributor) {
	server, err := api.NewServer(cfg, store, client, taskDistributor)
	if err != nil {
		zerolog.Fatal().Err(err).Msg("cannot create server")
	}

	// @contact.name aalug
	// @contact.url https://github.com/aalug
	// @contact.email a.a.gulczynski@gmail.com
	// @securityDefinitions.apikey ApiKeyAuth
	// @in header
	// @name Authorization
	// @description Use 'bearer {token}' without quotes.
	err = server.Start(cfg.ServerAddress)
	if err != nil {
		zerolog.Fatal().Err(err).Msg("cannot start the server")
	}
}

func runTaskProcessor(cfg config.Config, redisOpt asynq.RedisClientOpt, store db.Store) {
	emailSender := mail.NewHogSender(cfg.EmailSenderAddress)
	taskProcessor := worker.NewRedisTaskProcessor(redisOpt, store, emailSender)
	zerolog.Info().Msg("task processor started")
	err := taskProcessor.Start()
	if err != nil {
		zerolog.Fatal().Err(err).Msg("failed to start task processor")
	}
}
