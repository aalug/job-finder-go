// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.18.0

package db

import (
	"context"
)

type Querier interface {
	CreateCompany(ctx context.Context, arg CreateCompanyParams) (Company, error)
	DeleteCompany(ctx context.Context, id int32) error
	GetCompanyByID(ctx context.Context, id int32) (Company, error)
	GetCompanyByName(ctx context.Context, name string) (Company, error)
	UpdateCompany(ctx context.Context, arg UpdateCompanyParams) (Company, error)
}

var _ Querier = (*Queries)(nil)
