package api

import (
	"database/sql"
	"fmt"
	db "github.com/aalug/go-gin-job-search/db/sqlc"
	"github.com/aalug/go-gin-job-search/token"
	"github.com/gin-gonic/gin"
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
// @Failure 401 {object} ErrorResponse "Unauthorized. Only users can access"
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
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	ctx.JSON(http.StatusCreated, newJobApplicationResponse(jobApplication))
}
