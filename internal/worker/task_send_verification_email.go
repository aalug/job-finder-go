package worker

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/hibiken/asynq"
	"github.com/rs/zerolog/log"
)

const TaskSendVerificationEmail = "task:send_verification_email"

type PayloadSendVerificationEmail struct {
	Email string `json:"email"`
}

// DistributeTaskSendVerificationEmail distributes the task of sending a verification email.
func (distributor *RedisTaskDistributor) DistributeTaskSendVerificationEmail(
	ctx context.Context,
	payload *PayloadSendVerificationEmail,
	opts ...asynq.Option,
) error {
	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %w", err)
	}

	task := asynq.NewTask(TaskSendVerificationEmail, jsonPayload, opts...)
	info, err := distributor.client.EnqueueContext(ctx, task)
	if err != nil {
		return fmt.Errorf("failed to enqueue task: %w", err)
	}
	log.Info().Str("type", task.Type()).Bytes("payload", task.Payload()).
		Str("queue", info.Queue).Int("max_retry", info.MaxRetry).Msg("enqueued task")

	return nil
}

// ProcessTaskSendVerificationEmail processes the task of sending a verification email.
func (processor *RedisTaskProcessor) ProcessTaskSendVerificationEmail(ctx context.Context, task *asynq.Task) error {
	var payload PayloadSendVerificationEmail
	err := json.Unmarshal(task.Payload(), &payload)
	if err != nil {
		return fmt.Errorf("failed to unmarshal payload: %w", asynq.SkipRetry)
	}

	user, err := processor.store.GetUserByEmail(ctx, payload.Email)
	if err != nil {
		return fmt.Errorf("failed to get user: %w", err)
	}

	// send email to user

	log.Info().Str("type", task.Type()).Bytes("payload", task.Payload()).
		Str("email", user.Email).Msg("processed task")

	return nil
}
