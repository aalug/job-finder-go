package worker

import (
	"context"
	"github.com/hibiken/asynq"
)

type TaskDistributor interface {
	DistributeTaskSendVerificationEmail(
		ctx context.Context,
		payload *PayloadSendVerificationEmail,
		opts ...asynq.Option,
	) error
	DistributeTaskSendConfirmationEmail(
		ctx context.Context,
		payload *PayloadSendConfirmationEmail,
		opts ...asynq.Option,
	) error
}

type RedisTaskDistributor struct {
	client *asynq.Client
}

func NewRedisTaskDistributor(redisOpt asynq.RedisClientOpt) TaskDistributor {
	client := asynq.NewClient(redisOpt)
	return &RedisTaskDistributor{
		client: client,
	}
}
