package worker

import (
	"context"
	"github.com/aalug/job-finder-go/internal/config"
	db "github.com/aalug/job-finder-go/internal/db/sqlc"
	"github.com/aalug/job-finder-go/internal/mail"
	"github.com/hibiken/asynq"
	"github.com/rs/zerolog/log"
)

const (
	QueueCritical = "critical"
	QueueDefault  = "default"
)

type TaskProcessor interface {
	Start() error
	ProcessTaskSendVerificationEmail(ctx context.Context, task *asynq.Task) error
	ProcessTaskSendConfirmationEmail(ctx context.Context, task *asynq.Task) error
}

type RedisTaskProcessor struct {
	server      *asynq.Server
	store       db.Store
	emailSender mail.EmailSender
	config      config.Config
}

func NewRedisTaskProcessor(redisOpt asynq.RedisClientOpt, store db.Store, emailSender mail.EmailSender) TaskProcessor {
	server := asynq.NewServer(
		redisOpt,
		asynq.Config{
			Queues: map[string]int{
				QueueCritical: 10,
				QueueDefault:  5,
			},
			ErrorHandler: asynq.ErrorHandlerFunc(
				func(ctx context.Context, task *asynq.Task, err error) {
					// log error
					log.Error().Err(err).Str("type", task.Type()).
						Bytes("payload", task.Payload()).
						Msg("process task failed")
				}),
			Logger: NewLogger(),
		},
	)

	cfg, err := config.LoadConfig("../..")
	if err != nil {
		log.Fatal().Err(err).Msg("cannot load config")
	}

	return &RedisTaskProcessor{
		server:      server,
		store:       store,
		emailSender: emailSender,
		config:      cfg,
	}
}

// Start starts the processor
func (processor *RedisTaskProcessor) Start() error {
	mux := asynq.NewServeMux()

	mux.HandleFunc(TaskSendVerificationEmail, processor.ProcessTaskSendVerificationEmail)
	mux.HandleFunc(TaskSendConfirmationEmail, processor.ProcessTaskSendConfirmationEmail)

	return processor.server.Start(mux)
}
