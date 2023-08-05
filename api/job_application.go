package api

import (
	"database/sql"
	"fmt"
	db "github.com/aalug/go-gin-job-search/db/sqlc"
	"github.com/aalug/go-gin-job-search/token"
	"github.com/gin-gonic/gin"
	"github.com/lib/pq"
	"io"
	"net/http"
	"strconv"
	"time"
)

type jobApplicationResponse struct {
	ID        int32                `json:"id"`
	JobID     int32                `json:"job_id"`
	Message   string               `json:"message"`
	Status    db.ApplicationStatus `json:"status"`
	AppliedAt time.Time            `json:"applied_at"`
}

func newJobApplicationResponse(jobApplication db.JobApplication) jobApplicationResponse {
	return jobApplicationResponse{
		ID:        jobApplication.ID,
		JobID:     jobApplication.JobID,
		Message:   jobApplication.Message.String,
		Status:    jobApplication.Status,
		AppliedAt: jobApplication.AppliedAt,
	}
}

// @Schemes
// @Summary Create job application
// @Description Create a job application. Only users can access this endpoint.
// @Tags job applications
// @param cv formData file true "CV file"
// @param message formData string false "Message for the employer"
// @param job_id formData int true "Job ID"
// @Accept multipart/form-data
// @Produce json
// @Success 200 {object} jobApplicationResponse
// @Failure 400 {object} ErrorResponse "Invalid request body"
// @Failure 401 {object} ErrorResponse "Unauthorized. Only users can access, not employers."
// @Failure 500 {object} ErrorResponse "Any other error"
// @Router /job-applications [post]
// createJobApplication creates a new job application
func (server *Server) createJobApplication(ctx *gin.Context) {
	// check if the user is authenticated
	authPayload := ctx.MustGet(authorizationPayloadKey).(*token.Payload)
	authUser, err := server.store.GetUserByEmail(ctx, authPayload.Email)
	if err != nil {
		if err == sql.ErrNoRows {
			// but middleware did not stop the request, so it had to be made by the employer
			ctx.JSON(http.StatusUnauthorized, errorResponse(onlyUsersAccessError))
			return
		}

		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	// get the CV file
	file, header, err := ctx.Request.FormFile("cv")
	if err != nil || header == nil {
		err = fmt.Errorf("valid CV file is required: %w", err)
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}
	defer file.Close()

	// Read the file data and convert it to a byte slice
	cvData, err := io.ReadAll(file)
	if err != nil {
		err = fmt.Errorf("failed to read the CV file: %w", err)
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	// get the message and the jobID
	message := ctx.Request.FormValue("message")
	jobIDStr := ctx.Request.FormValue("job_id")

	// Validate the jobID
	jobID, err := strconv.Atoi(jobIDStr)
	if err != nil || jobID <= 0 {
		err = fmt.Errorf("invalid job ID. Please provide a valid positive integer job ID")
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	// create job application in the database
	params := db.CreateJobApplicationParams{
		UserID: authUser.ID,
		JobID:  int32(jobID),
		Message: sql.NullString{
			String: message,
			Valid:  len(message) > 0,
		},
		Cv: cvData,
	}

	jobApplication, err := server.store.CreateJobApplication(ctx, params)
	if err != nil {
		if pqErr, ok := err.(*pq.Error); ok {
			switch pqErr.Code.Name() {
			case "unique_violation":
				err := fmt.Errorf("user with ID %d has already applied for this job", authUser.ID)
				ctx.JSON(http.StatusForbidden, errorResponse(err))
				return
			}
		}
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	ctx.JSON(http.StatusCreated, newJobApplicationResponse(jobApplication))
}

type getJobApplicationForUserRequest struct {
	ID int32 `uri:"id" binding:"required,min=1"`
}

type getJobApplicationForUserResponse struct {
	ApplicationID      int32                `json:"application_id"`
	JobID              int32                `json:"job_id"`
	JobTitle           string               `json:"job_title"`
	CompanyName        string               `json:"company_name"`
	ApplicationStatus  db.ApplicationStatus `json:"application_status"`
	ApplicationDate    time.Time            `json:"application_date"`
	ApplicationMessage string               `json:"application_message"`
	UserCv             []byte               `json:"user_cv"`
	UserID             int32                `json:"user_id"`
}

// @Schemes
// @Summary Get job application for user
// @Description Get job application for a user. Only users can access this endpoint. It returns different details than getJobApplicationForEmployer.
// @Tags job applications
// @param id path int true "job application ID"
// @Produce json
// @Success 200 {object} getJobApplicationForUserResponse
// @Failure 400 {object} ErrorResponse "Invalid ID"
// @Failure 401 {object} ErrorResponse "Unauthorized. Only users can access, not employers."
// @Failure 403 {object} ErrorResponse "Only the applicant (the owner) of the job application can access this endpoint."
// @Failure 500 {object} ErrorResponse "Any other error"
// @Router /job-applications/user/{id} [get]
// getJobApplicationForUser gets a job application for a user.
// Only users can access this endpoint and only the applicant (the owner)
// of the job application will receive the success response.
// It also returns different details than getJobApplicationForEmployer
// (suitable for the user needs)
func (server *Server) getJobApplicationForUser(ctx *gin.Context) {
	var request getJobApplicationForUserRequest
	if err := ctx.ShouldBindUri(&request); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	// check if the user is authenticated (and is a user, not employer)
	authPayload := ctx.MustGet(authorizationPayloadKey).(*token.Payload)
	authUser, err := server.store.GetUserByEmail(ctx, authPayload.Email)
	if err != nil {
		if err == sql.ErrNoRows {
			// but middleware did not stop the request, so it had to be made by an employer
			ctx.JSON(http.StatusUnauthorized, errorResponse(onlyUsersAccessError))
			return
		}

		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	// get the job application from the database
	jobApplication, err := server.store.GetJobApplicationForUser(ctx, request.ID)
	if err != nil {
		if err == sql.ErrNoRows {
			err = fmt.Errorf("job application with ID %d does not exist", request.ID)
			ctx.JSON(http.StatusNotFound, errorResponse(err))
			return
		}

		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	// check if the authenticated user is the applicant
	if authUser.ID != jobApplication.UserID {
		err = fmt.Errorf("user with ID %d is not the applicant of this job application", authUser.ID)
		ctx.JSON(http.StatusForbidden, errorResponse(err))
		return
	}

	// TODO: for now, CV is sent as []byte, but later it will be hosted in a file server
	res := getJobApplicationForUserResponse{
		ApplicationID:     jobApplication.ApplicationID,
		JobID:             jobApplication.JobID,
		JobTitle:          jobApplication.JobTitle,
		CompanyName:       jobApplication.CompanyName,
		ApplicationStatus: jobApplication.ApplicationStatus,
		ApplicationDate:   jobApplication.ApplicationDate,
		UserCv:            jobApplication.UserCv,
	}
	if jobApplication.ApplicationMessage.Valid {
		res.ApplicationMessage = jobApplication.ApplicationMessage.String
	}

	ctx.JSON(http.StatusOK, res)
}

type getJobApplicationForEmployerRequest struct {
	ID int32 `uri:"id" binding:"required,min=1"`
}

type getJobApplicationForEmployerResponse struct {
	ApplicationID      int32                `json:"application_id"`
	JobTitle           string               `json:"job_title"`
	JobID              int32                `json:"job_id"`
	ApplicationStatus  db.ApplicationStatus `json:"application_status"`
	ApplicationDate    time.Time            `json:"application_date"`
	ApplicationMessage string               `json:"application_message"`
	UserCv             []byte               `json:"user_cv"`
	UserID             int32                `json:"user_id"`
	UserEmail          string               `json:"user_email"`
	UserFullName       string               `json:"user_full_name"`
	UserLocation       string               `json:"user_location"`
}

// @Schemes
// @Summary Get job application for employer
// @Description Get job application for an employer. Only employers can access this endpoint. It returns different details than getJobApplicationForUser.
// @Tags job applications
// @param id path int true "job application ID"
// @Produce json
// @Success 200 {object} getJobApplicationForEmployerResponse
// @Failure 400 {object} ErrorResponse "Invalid ID"
// @Failure 401 {object} ErrorResponse "Unauthorized. Only employers can access, not users."
// @Failure 403 {object} ErrorResponse "Only an employer that is part of the company that created this application can access this endpoint.
// @Failure 500 {object} ErrorResponse "Any other error"
// @Router /job-applications/employer/{id} [get]
// getJobApplicationForEmployer gets a job application for an employer.
// Only employers can access this endpoint and only employers that are part
// of the company will receive the success response.
// It also returns different details than getJobApplicationForUser
// (suitable for the employer needs)
func (server *Server) getJobApplicationForEmployer(ctx *gin.Context) {
	var request getJobApplicationForEmployerRequest
	if err := ctx.ShouldBindUri(&request); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	// check if the employer is authenticated (and is an employer, not user)
	authPayload := ctx.MustGet(authorizationPayloadKey).(*token.Payload)
	authEmployer, err := server.store.GetEmployerByEmail(ctx, authPayload.Email)
	if err != nil {
		if err == sql.ErrNoRows {
			// but middleware did not stop the request, so it had to be made by a user
			ctx.JSON(http.StatusUnauthorized, errorResponse(onlyUsersAccessError))
			return
		}

		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	// get the job application from the database
	jobApplication, err := server.store.GetJobApplicationForEmployer(ctx, request.ID)
	if err != nil {
		if err == sql.ErrNoRows {
			err = fmt.Errorf("job application with ID %d does not exist", request.ID)
			ctx.JSON(http.StatusNotFound, errorResponse(err))
			return
		}

		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	// check if the authenticated employer is part of the company
	// that created this application
	companyID, err := server.store.GetCompanyIDOfJob(ctx, jobApplication.JobID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	if companyID != authEmployer.CompanyID {
		err = fmt.Errorf("employer with ID %d is not part of the company that created this job", authEmployer.ID)
		ctx.JSON(http.StatusForbidden, errorResponse(err))
		return
	}

	res := getJobApplicationForEmployerResponse{
		ApplicationID:     jobApplication.ApplicationID,
		JobTitle:          jobApplication.JobTitle,
		JobID:             jobApplication.JobID,
		ApplicationStatus: jobApplication.ApplicationStatus,
		ApplicationDate:   jobApplication.ApplicationDate,
		UserCv:            jobApplication.UserCv,
		UserID:            jobApplication.UserID,
		UserEmail:         jobApplication.UserEmail,
		UserFullName:      jobApplication.UserFullName,
		UserLocation:      jobApplication.UserLocation,
	}

	// TODO: for now, CV is sent as []byte, but later it will be hosted in a file server

	if jobApplication.ApplicationMessage.Valid {
		res.ApplicationMessage = jobApplication.ApplicationMessage.String
	}

	// if the application status was 'Applied', change it to `Seen`
	if jobApplication.ApplicationStatus == db.ApplicationStatusApplied {
		err = server.store.UpdateJobApplicationStatus(ctx, db.UpdateJobApplicationStatusParams{
			ID:     jobApplication.ApplicationID,
			Status: db.ApplicationStatusSeen,
		})

		if err != nil {
			ctx.JSON(http.StatusInternalServerError, errorResponse(err))
			return
		}
	}

	ctx.JSON(http.StatusOK, res)
}
