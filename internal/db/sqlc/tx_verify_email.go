package db

import (
	"context"
)

type VerifyEmailTxParams struct {
	ID         int64
	SecretCode string
}

type VerifyUserEmailResult struct {
	User        User
	VerifyEmail VerifyEmail
}

// VerifyUserEmailTx verify user email transaction
func (store *SQLStore) VerifyUserEmailTx(ctx context.Context, arg VerifyEmailTxParams) (VerifyUserEmailResult, error) {
	var result VerifyUserEmailResult

	err := store.ExecTx(ctx, func(q *Queries) error {
		var err error

		result.VerifyEmail, err = q.UpdateVerifyEmail(ctx, UpdateVerifyEmailParams{
			ID:         arg.ID,
			SecretCode: arg.SecretCode,
		})
		if err != nil {
			return err
		}

		result.User, err = q.VerifyUserEmail(ctx, result.VerifyEmail.Email)
		return err
	})

	return result, err
}

type VerifyEmployerEmailResult struct {
	Employer    Employer
	VerifyEmail VerifyEmail
}

// VerifyEmployerEmailTx verify employer email transaction
func (store *SQLStore) VerifyEmployerEmailTx(ctx context.Context, arg VerifyEmailTxParams) (VerifyEmployerEmailResult, error) {
	var result VerifyEmployerEmailResult

	err := store.ExecTx(ctx, func(q *Queries) error {
		var err error

		result.VerifyEmail, err = q.UpdateVerifyEmail(ctx, UpdateVerifyEmailParams{
			ID:         arg.ID,
			SecretCode: arg.SecretCode,
		})
		if err != nil {
			return err
		}

		result.Employer, err = q.VerifyEmployerEmail(ctx, result.VerifyEmail.Email)
		return err
	})

	return result, err
}
