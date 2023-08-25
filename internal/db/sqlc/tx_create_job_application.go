package db

import "context"

type CreateJobApplicationTxParams struct {
	CreateJobApplicationParams
	AfterCreate func(jobApplication JobApplication) error
}

type CreateJobApplicationTxResult struct {
	JobApplication JobApplication
}

func (store *SQLStore) CreateJobApplicationTx(ctx context.Context, arg CreateJobApplicationTxParams) (CreateJobApplicationTxResult, error) {
	var result CreateJobApplicationTxResult

	err := store.ExecTx(ctx, func(q *Queries) error {
		var err error

		result.JobApplication, err = q.CreateJobApplication(ctx, arg.CreateJobApplicationParams)
		if err != nil {
			return err
		}

		return arg.AfterCreate(result.JobApplication)
	})

	return result, err
}
