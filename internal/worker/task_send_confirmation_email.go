package worker

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/aalug/job-finder-go/internal/mail"
	"github.com/hibiken/asynq"
	"github.com/rs/zerolog/log"
)

const TaskSendConfirmationEmail = "task:send_confirmation_email"

type PayloadSendConfirmationEmail struct {
	Email       string `json:"email"`
	FullName    string `json:"full_name"`
	Position    string `json:"position"`
	CompanyName string `json:"company_name"`
}

// DistributeTaskSendConfirmationEmail distributes the task of sending a confirmation email.
func (distributor *RedisTaskDistributor) DistributeTaskSendConfirmationEmail(
	ctx context.Context,
	payload *PayloadSendConfirmationEmail,
	opts ...asynq.Option,
) error {
	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %w", err)
	}

	task := asynq.NewTask(TaskSendConfirmationEmail, jsonPayload, opts...)
	info, err := distributor.client.EnqueueContext(ctx, task)
	if err != nil {
		return fmt.Errorf("failed to enqueue task: %w", err)
	}
	log.Info().Str("type", task.Type()).Bytes("payload", task.Payload()).
		Str("queue", info.Queue).Int("max_retry", info.MaxRetry).Msg("enqueued task")

	return nil
}

// ProcessTaskSendConfirmationEmail processes the task of sending a confirmation email.
func (processor *RedisTaskProcessor) ProcessTaskSendConfirmationEmail(ctx context.Context, task *asynq.Task) error {
	var payload PayloadSendConfirmationEmail
	err := json.Unmarshal(task.Payload(), &payload)
	if err != nil {
		return fmt.Errorf("failed to unmarshal payload: %w", asynq.SkipRetry)
	}

	// send email to user to confirm creating a job application
	content := fmt.Sprintf(`
		<h3>Hello %s</h3><br>
		<p class="message">
		We would like to confirm the successful submission of your job application for the 
		%s position in %s through the Go Job Search.
		<br><br>
		Best regards,
		<strong>Go Job Search</strong>
		</p>
		`, payload.FullName, payload.Position, payload.CompanyName)
	err = processor.emailSender.SendEmail(mail.Data{
		To:       []string{payload.Email},
		Subject:  fmt.Sprintf("Job Application Confirmation - %s", payload.Position),
		Content:  content,
		Template: "confirmation_email.html",
	})
	if err != nil {
		return fmt.Errorf("failed to send email: %w", err)
	}

	log.Info().Str("type", task.Type()).Bytes("payload", task.Payload()).
		Str("email", payload.Email).Msg("processed task")

	return nil
}
