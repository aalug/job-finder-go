package worker

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	db "github.com/aalug/go-gin-job-search/internal/db/sqlc"
	"github.com/aalug/go-gin-job-search/internal/mail"
	"github.com/aalug/go-gin-job-search/pkg/utils"
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
// It works for both employers and users.
func (processor *RedisTaskProcessor) ProcessTaskSendVerificationEmail(ctx context.Context, task *asynq.Task) error {
	var payload PayloadSendVerificationEmail
	err := json.Unmarshal(task.Payload(), &payload)
	if err != nil {
		return fmt.Errorf("failed to unmarshal payload: %w", asynq.SkipRetry)
	}

	var email string
	var fullName string
	user, err := processor.store.GetUserByEmail(ctx, payload.Email)
	if err != nil {
		if err == sql.ErrNoRows {
			// it might be an employer
			employer, err := processor.store.GetEmployerByEmail(ctx, payload.Email)
			if err != nil {
				return fmt.Errorf("failed to get user: %w", err)
			}
			email = employer.Email
			fullName = employer.FullName
		} else {
			return fmt.Errorf("failed to get user: %w", err)
		}

	} else {
		email = user.Email
		fullName = user.FullName
	}

	// create verify email in the database
	verifyEmail, err := processor.store.CreateVerifyEmail(ctx, db.CreateVerifyEmailParams{
		Email:      email,
		SecretCode: utils.RandomString(32),
	})
	if err != nil {
		return fmt.Errorf("failed to create verify email in the db: %w", err)
	}

	// send email to user to verify email
	verifyUrl := fmt.Sprintf("/%s%s/employers/verify-email?id=%d&code=%s",
		processor.config.ServerAddress, processor.config.BaseUrl, verifyEmail.ID, verifyEmail.SecretCode)
	content := fmt.Sprintf(`
		<h3>Hello %s</h3><br>
		<p class="message">
		Please click the link below to verify your email address:
		</p>
		<a class="button" href="%s">Verify Email</a>
		`, fullName, verifyUrl)
	err = processor.emailSender.SendEmail(mail.Data{
		To:       []string{email},
		Subject:  "Welcome to Go Job Search!",
		Content:  content,
		Template: "verification_email.html",
	})
	if err != nil {
		return fmt.Errorf("failed to send email: %w", err)
	}

	log.Info().Str("type", task.Type()).Bytes("payload", task.Payload()).
		Str("email", user.Email).Msg("processed task")

	return nil
}
