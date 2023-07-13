package api

import (
	"fmt"
	db "github.com/aalug/go-gin-job-search/db/sqlc"
	"github.com/aalug/go-gin-job-search/utils"
	"github.com/gin-gonic/gin"
	"github.com/lib/pq"
	"net/http"
	"time"
)

type createEmployerRequest struct {
	FullName        string `json:"full_name" binding:"required"`
	Email           string `json:"email" binding:"required,email"`
	Password        string `json:"password" binding:"required,min=6"`
	CompanyName     string `json:"company_name" binding:"required"`
	CompanyIndustry string `json:"company_industry" binding:"required"`
	CompanyLocation string `json:"company_location" binding:"required"`
}

type employerResponse struct {
	EmployerID        int32     `json:"employer_id"`
	FullName          string    `json:"full_name"`
	Email             string    `json:"email"`
	EmployerCreatedAt time.Time `json:"employer_created_at"`
	CompanyID         int32     `json:"company_id"`
	CompanyName       string    `json:"company_name"`
	CompanyIndustry   string    `json:"company_industry"`
	CompanyLocation   string    `json:"company_location"`
}

// newEmployerResponse creates a new employer response from a db.Employer and db.Company
func newEmployerResponse(employer db.Employer, company db.Company) employerResponse {
	return employerResponse{
		EmployerID:        employer.ID,
		FullName:          employer.FullName,
		Email:             employer.Email,
		EmployerCreatedAt: employer.CreatedAt,
		CompanyID:         company.ID,
		CompanyName:       company.Name,
		CompanyIndustry:   company.Industry,
		CompanyLocation:   company.Location,
	}
}

// createEmployer handles creation of an employer
func (server *Server) createEmployer(ctx *gin.Context) {
	var request createEmployerRequest
	if err := ctx.ShouldBindJSON(&request); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	// create a company
	companyParams := db.CreateCompanyParams{
		Name:     request.CompanyName,
		Industry: request.CompanyIndustry,
		Location: request.CompanyLocation,
	}

	company, err := server.store.CreateCompany(ctx, companyParams)
	if err != nil {
		if pqErr, ok := err.(*pq.Error); ok {
			switch pqErr.Code.Name() {
			case "unique_violation":
				err := fmt.Errorf("company with this name already exists")
				ctx.JSON(http.StatusForbidden, errorResponse(err))
				return
			}
		}
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	// hash password
	hashedPassword, err := utils.HashPassword(request.Password)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	// Create an employer
	employerParams := db.CreateEmployerParams{
		CompanyID:      company.ID,
		FullName:       request.FullName,
		Email:          request.Email,
		HashedPassword: hashedPassword,
	}

	employer, err := server.store.CreateEmployer(ctx, employerParams)
	if err != nil {
		if pqErr, ok := err.(*pq.Error); ok {
			switch pqErr.Code.Name() {
			case "unique_violation":
				err := fmt.Errorf("employer with this email already exists")
				ctx.JSON(http.StatusForbidden, errorResponse(err))
				return
			}
		}
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	ctx.JSON(http.StatusCreated, newEmployerResponse(employer, company))
}
