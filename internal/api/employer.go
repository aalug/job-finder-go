package api

import (
	"database/sql"
	"encoding/json"
	"fmt"
	db "github.com/aalug/go-gin-job-search/internal/db/sqlc"
	"github.com/aalug/go-gin-job-search/pkg/token"
	"github.com/aalug/go-gin-job-search/pkg/utils"
	"github.com/aalug/go-gin-job-search/pkg/validation"
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

// @Schemes
// @Summary Create employer
// @Description Create a new employer
// @Tags employers
// @Accept json
// @Produce json
// @param CreateEmployerRequest body createEmployerRequest true "Employer and company details"
// @Success 201 {object} employerResponse
// @Failure 400 {object} ErrorResponse "Invalid request body"
// @Failure 403 {object} ErrorResponse "Company with given name or employer with given email already exists"
// @Failure 500 {object} ErrorResponse "Any other error"
// @Router /employers [post]
// createEmployer handles creating a new employer
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

type loginEmployerRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6"`
}

type loginEmployerResponse struct {
	AccessToken string           `json:"access_token"`
	Employer    employerResponse `json:"employer"`
}

// @Schemes
// @Summary Login employer
// @Description Login an employer
// @Tags employers
// @Accept json
// @Produce json
// @param LoginEmployerRequest body loginEmployerRequest true "Employer credentials"
// @Success 200 {object} loginEmployerResponse
// @Failure 400 {object} ErrorResponse "Invalid request body"
// @Failure 404 {object} ErrorResponse "Employer with given email or company with given id does not exist"
// @Failure 401 {object} ErrorResponse "Incorrect password"
// @Failure 500 {object} ErrorResponse "Any other error"
// @Router /employers/login [post]
// loginEmployer handles login of an employer
func (server *Server) loginEmployer(ctx *gin.Context) {
	var request loginEmployerRequest
	if err := ctx.ShouldBindJSON(&request); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	// get the employer
	employer, err := server.store.GetEmployerByEmail(ctx, request.Email)
	if err != nil {
		if err == sql.ErrNoRows {
			err = fmt.Errorf("employer with this email does not exist")
			ctx.JSON(http.StatusNotFound, errorResponse(err))
			return
		}
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	// check password
	err = utils.CheckPassword(request.Password, employer.HashedPassword)
	if err != nil {
		err = fmt.Errorf("incorrect password")
		ctx.JSON(http.StatusUnauthorized, errorResponse(err))
		return
	}

	// create access token
	accessToken, err := server.tokenMaker.CreateToken(employer.Email, server.config.AccessTokenDuration)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	// get employers company
	company, err := server.store.GetCompanyByID(ctx, employer.CompanyID)
	if err != nil {
		if err == sql.ErrNoRows {
			err = fmt.Errorf("company with this id does not exist")
			ctx.JSON(http.StatusNotFound, errorResponse(err))
			return
		}

		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	res := loginEmployerResponse{
		AccessToken: accessToken,
		Employer:    newEmployerResponse(employer, company),
	}

	ctx.JSON(http.StatusOK, res)
}

// @Schemes
// @Summary Get employer
// @Description Get the details of the authenticated employer
// @Tags employers
// @Produce json
// @Success 200 {object} employerResponse
// @Failure 401 {object} ErrorResponse "Only employers can access this endpoint."
// @Failure 500 {object} ErrorResponse "Internal error"
// @Security ApiKeyAuth
// @Router /employers [get]
// getEmployer get details of the authenticated employer
func (server *Server) getEmployer(ctx *gin.Context) {
	authPayload := ctx.MustGet(authorizationPayloadKey).(*token.Payload)
	authEmployer, err := server.store.GetEmployerByEmail(ctx, authPayload.Email)
	if err != nil {
		if err == sql.ErrNoRows {
			// but middleware did not stop the request, so we assume
			// that the request was made by a user
			ctx.JSON(http.StatusUnauthorized, errorResponse(onlyEmployersAccessError))
			return
		}

		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	company, err := server.store.GetCompanyByID(ctx, authEmployer.CompanyID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	ctx.JSON(http.StatusOK, newEmployerResponse(authEmployer, company))
}

type updateEmployerRequest struct {
	FullName        string `json:"full_name"`
	Email           string `json:"email"`
	CompanyName     string `json:"company_name"`
	CompanyIndustry string `json:"company_industry"`
	CompanyLocation string `json:"company_location"`
}

// @Schemes
// @Summary Update employer
// @Description Update the details of the authenticated employer
// @Tags employers
// @Accept json
// @Produce json
// @param UpdateEmployerRequest body updateEmployerRequest true "Employer details to update"
// @Success 200 {object} employerResponse
// @Failure 400 {object} ErrorResponse "Invalid request body"
// @Failure 401 {object} ErrorResponse "Only employers can access this endpoint."
// @Failure 500 {object} ErrorResponse "Any other error"
// @Security ApiKeyAuth
// @Router /employers [patch]
// updateEmployer handles update of an employer details
func (server *Server) updateEmployer(ctx *gin.Context) {
	var request updateEmployerRequest
	err := json.NewDecoder(ctx.Request.Body).Decode(&request)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	if request.Email != "" {
		if err := validation.ValidateEmail(request.Email); err != nil {
			ctx.JSON(http.StatusBadRequest, errorResponse(err))
			return
		}
	}

	authPayload := ctx.MustGet(authorizationPayloadKey).(*token.Payload)
	authEmployer, err := server.store.GetEmployerByEmail(ctx, authPayload.Email)
	if err != nil {
		if err == sql.ErrNoRows {
			// but middleware did not stop the request, so we assume
			// that the request was made by a user
			ctx.JSON(http.StatusUnauthorized, errorResponse(onlyEmployersAccessError))
			return
		}

		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	company, err := server.store.GetCompanyByID(ctx, authEmployer.CompanyID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	// update the company details
	companyParams := db.UpdateCompanyParams{
		ID:       company.ID,
		Name:     company.Name,
		Industry: company.Industry,
		Location: company.Location,
	}

	shouldUpdateCompany := false
	if request.CompanyName != "" {
		companyParams.Name = request.CompanyName
		shouldUpdateCompany = true
	}
	if request.CompanyIndustry != "" {
		companyParams.Industry = request.CompanyIndustry
		shouldUpdateCompany = true
	}
	if request.CompanyLocation != "" {
		companyParams.Location = request.CompanyLocation
		shouldUpdateCompany = true
	}

	if shouldUpdateCompany {
		// Update the company
		company, err = server.store.UpdateCompany(ctx, companyParams)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, errorResponse(err))
			return
		}

		employerParams := db.UpdateEmployerParams{
			ID:        authEmployer.ID,
			CompanyID: authEmployer.CompanyID,
			FullName:  authEmployer.FullName,
			Email:     authEmployer.Email,
		}

		shouldUpdateEmployer := false
		if request.Email != "" {
			employerParams.Email = request.Email
			shouldUpdateEmployer = true
		}
		if request.FullName != "" {
			employerParams.FullName = request.FullName
			shouldUpdateEmployer = true
		}

		if shouldUpdateEmployer {
			// Update the employer
			authEmployer, err = server.store.UpdateEmployer(ctx, employerParams)
			if err != nil {
				ctx.JSON(http.StatusInternalServerError, errorResponse(err))
				return
			}
		}
	}

	ctx.JSON(http.StatusOK, newEmployerResponse(authEmployer, company))
}

type updateEmployerPasswordRequest struct {
	OldPassword string `json:"old_password" binding:"required,min=6"`
	NewPassword string `json:"new_password" binding:"required,min=6"`
}

type updateEmployerPasswordResponse struct {
	Message string `json:"message"`
}

// @Schemes
// @Summary Update employer password
// @Description Update/change logged-in employer password
// @Tags employers
// @Accept json
// @Produce json
// @param UpdateEmployerPasswordRequest body updateEmployerPasswordRequest true "Employer old and new password"
// @Success 200 {object} updateEmployerPasswordResponse
// @Failure 400 {object} ErrorResponse "Invalid request body"
// @Failure 401 {object} ErrorResponse "Incorrect password or request made a user, not employer."
// @Failure 500 {object} ErrorResponse "Any other error"
// @Security ApiKeyAuth
// @Router /employers/password [patch]
// updateEmployerPassword handles user password update
func (server *Server) updateEmployerPassword(ctx *gin.Context) {
	var request updateEmployerPasswordRequest
	if err := ctx.ShouldBindJSON(&request); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	authPayload := ctx.MustGet(authorizationPayloadKey).(*token.Payload)
	authEmployer, err := server.store.GetEmployerByEmail(ctx, authPayload.Email)
	if err != nil {
		if err == sql.ErrNoRows {
			// but middleware did not stop the request, so we assume
			// that the request was made by a user
			ctx.JSON(http.StatusUnauthorized, errorResponse(onlyEmployersAccessError))
			return
		}

		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	err = utils.CheckPassword(request.OldPassword, authEmployer.HashedPassword)
	if err != nil {
		err = fmt.Errorf("incorrect password")
		ctx.JSON(http.StatusUnauthorized, errorResponse(err))
		return
	}

	hashedPassword, err := utils.HashPassword(request.NewPassword)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	params := db.UpdateEmployerPasswordParams{
		ID:             authEmployer.ID,
		HashedPassword: hashedPassword,
	}

	err = server.store.UpdateEmployerPassword(ctx, params)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	ctx.JSON(http.StatusOK, updateEmployerPasswordResponse{"password updated successfully"})
}

// @Schemes
// @Summary Delete employer
// @Description Delete the logged-in employer
// @Tags employers
// @Success 204 {null} null
// @Failure 401 {object} ErrorResponse "Only employers can access this endpoint."
// @Failure 500 {object} ErrorResponse "Any other error"
// @Security ApiKeyAuth
// @Router /employers [delete]
// deleteEmployer handles deleting employer
func (server *Server) deleteEmployer(ctx *gin.Context) {
	authPayload := ctx.MustGet(authorizationPayloadKey).(*token.Payload)
	authEmployer, err := server.store.GetEmployerByEmail(ctx, authPayload.Email)
	if err != nil {
		if err == sql.ErrNoRows {
			// but middleware did not stop the request, so we assume
			// that the request was made by a user
			ctx.JSON(http.StatusUnauthorized, errorResponse(onlyEmployersAccessError))
			return
		}

		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	// delete the company
	err = server.store.DeleteCompany(ctx, authEmployer.CompanyID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	// delete the employer
	err = server.store.DeleteEmployer(ctx, authEmployer.ID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	ctx.JSON(http.StatusNoContent, nil)
}

type getUserAsEmployerRequest struct {
	Email string `uri:"email" binding:"required,email"`
}

// @Schemes
// @Summary Get user as employer
// @Description Get a user as employer. Returns user details and skills. Only employers can access this endpoint.
// @Tags employers
// @Success 200 {object} userResponse
// @Failure 400 {object} ErrorResponse "Invalid email in uri."
// @Failure 401 {object} ErrorResponse "Only employers can access this endpoint."
// @Failure 404 {object} ErrorResponse "User with given email does not exist."
// @Failure 500 {object} ErrorResponse "Any other error."
// @Security ApiKeyAuth
// @Router /employers/user-details/{email} [get]
// getUserAsEmployer get user details as employer.
func (server *Server) getUserAsEmployer(ctx *gin.Context) {
	var request getUserAsEmployerRequest
	if err := ctx.ShouldBindUri(&request); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	// authenticate the employer
	authPayload := ctx.MustGet(authorizationPayloadKey).(*token.Payload)
	_, err := server.store.GetEmployerByEmail(ctx, authPayload.Email)
	if err != nil {
		if err == sql.ErrNoRows {
			// but middleware did not stop the request, so we assume
			// that the request was made by a user
			ctx.JSON(http.StatusUnauthorized, errorResponse(onlyEmployersAccessError))
			return
		}

		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	user, userSkills, err := server.store.GetUserDetailsByEmail(ctx, request.Email)
	if err != nil {
		if err == sql.ErrNoRows {
			err = fmt.Errorf("user with email %s does not exist", request.Email)
			ctx.JSON(http.StatusNotFound, errorResponse(err))
			return
		}

		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	ctx.JSON(http.StatusOK, newUserResponse(user, userSkills))
}

type getEmployerAndCompanyDetailsRequest struct {
	Email string `uri:"email" binding:"required,email"`
}

// @Schemes
// @Summary Get employer and company details as user
// @Description Get employer and company details as user. Does not require authentication.
// @Tags employers
// @Success 200 {object} db.GetEmployerAndCompanyDetailsRow
// @Failure 400 {object} ErrorResponse "Invalid email in uri."
// @Failure 404 {object} ErrorResponse "Employer with given email does not exist."
// @Failure 500 {object} ErrorResponse "Any other error."
// @Router /users/employer-company-details/{email} [get]
// getEmployerAsUser get employer and company details as user.
func (server *Server) getEmployerAndCompanyDetails(ctx *gin.Context) {
	var request getEmployerAndCompanyDetailsRequest
	if err := ctx.ShouldBindUri(&request); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	// get employer and company information
	details, err := server.store.GetEmployerAndCompanyDetails(ctx, request.Email)
	if err != nil {
		if err == sql.ErrNoRows {
			err = fmt.Errorf("employer with email %s does not exist", request.Email)
			ctx.JSON(http.StatusNotFound, errorResponse(err))
			return
		}

		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	ctx.JSON(http.StatusOK, details)
}
