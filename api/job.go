package api

import (
	db "github.com/aalug/go-gin-job-search/db/sqlc"
	"github.com/aalug/go-gin-job-search/token"
	"github.com/gin-gonic/gin"
	"net/http"
)

type jobResponse struct {
	Title          string                       `json:"title"`
	Description    string                       `json:"description"`
	Industry       string                       `json:"industry"`
	Location       string                       `json:"location"`
	SalaryMin      int32                        `json:"salary_min"`
	SalaryMax      int32                        `json:"salary_max"`
	Requirements   string                       `json:"requirements"`
	RequiredSkills []db.ListJobSkillsByJobIDRow `json:"required_skills"`
}

// newJobResponse creates a job response from a db.Job and db.ListJobSkillsByJobIDRow
func newJobResponse(job db.Job, skills []db.ListJobSkillsByJobIDRow) jobResponse {
	return jobResponse{
		Title:          job.Title,
		Description:    job.Description,
		Industry:       job.Industry,
		Location:       job.Location,
		SalaryMin:      job.SalaryMin,
		SalaryMax:      job.SalaryMax,
		Requirements:   job.Requirements,
		RequiredSkills: skills,
	}
}

type createJobRequest struct {
	Title          string   `json:"title" binding:"required"`
	Description    string   `json:"description" binding:"required"`
	Industry       string   `json:"industry" binding:"required"`
	Location       string   `json:"location" binding:"required"`
	SalaryMin      int32    `json:"salary_min" binding:"required,min=0"`
	SalaryMax      int32    `json:"salary_max" binding:"required,min=0"`
	Requirements   string   `json:"requirements" binding:"required"`
	RequiredSkills []string `json:"required_skills" binding:"required"`
}

// createJob handles creating a job posting - job with job skills
func (server *Server) createJob(ctx *gin.Context) {
	var request createJobRequest
	if err := ctx.ShouldBindJSON(&request); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	authPayload := ctx.MustGet(authorizationPayloadKey).(*token.Payload)
	authEmployer, err := server.store.GetEmployerByEmail(ctx, authPayload.Email)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	// create job
	params := db.CreateJobParams{
		Title:        request.Title,
		Industry:     request.Industry,
		CompanyID:    authEmployer.CompanyID,
		Description:  request.Description,
		Location:     request.Location,
		SalaryMin:    request.SalaryMin,
		SalaryMax:    request.SalaryMax,
		Requirements: request.Requirements,
	}

	job, err := server.store.CreateJob(ctx, params)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	// create job skills
	err = server.store.CreateMultipleJobSkills(ctx, request.RequiredSkills, job.ID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	listJobSkillsParams := db.ListJobSkillsByJobIDParams{
		JobID:  job.ID,
		Limit:  10,
		Offset: 0,
	}

	jobSkills, err := server.store.ListJobSkillsByJobID(ctx, listJobSkillsParams)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	ctx.JSON(http.StatusCreated, newJobResponse(job, jobSkills))
}
