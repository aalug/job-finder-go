package db

import "context"

type CreateEmployerTxParams struct {
	CreateEmployerParams
	AfterCreate func(employer Employer) error
}

type CreateEmployerTxResult struct {
	Employer Employer
}

func (store *SQLStore) CreateEmployerTx(ctx context.Context, arg CreateEmployerTxParams) (CreateEmployerTxResult, error) {
	var result CreateEmployerTxResult

	err := store.ExecTx(ctx, func(q *Queries) error {
		var err error

		result.Employer, err = q.CreateEmployer(ctx, arg.CreateEmployerParams)
		if err != nil {
			return err
		}

		return arg.AfterCreate(result.Employer)
	})

	return result, err
}
