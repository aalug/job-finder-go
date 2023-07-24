package api

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	db "github.com/aalug/go-gin-job-search/db/sqlc"
	"github.com/aalug/go-gin-job-search/token"
	"github.com/gin-gonic/gin"
	"net/http"
)

var (
	onlyEmployersAccessError = errors.New("only employers can access this endpoint")
	onlyUsersAccessError     = errors.New("only users can access this endpoint")
	jobOwnershipError        = errors.New("job does not belong to this employer")
	salaryRangeError         = errors.New("salary min cannot be greater than salary max")
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

	if request.SalaryMin > request.SalaryMax {
		ctx.JSON(http.StatusBadRequest, errorResponse(salaryRangeError))
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

type deleteJobRequest struct {
	ID int32 `uri:"id" binding:"required,min=1"`
}

// deleteJob handles deleting a job posting
func (server *Server) deleteJob(ctx *gin.Context) {
	var request deleteJobRequest
	if err := ctx.ShouldBindUri(&request); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	// get employer that is making the request
	authPayload := ctx.MustGet(authorizationPayloadKey).(*token.Payload)
	authEmployer, err := server.store.GetEmployerByEmail(ctx, authPayload.Email)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	// get job that is being deleted
	job, err := server.store.GetJob(ctx, request.ID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	// check if job is owned by the employer
	if job.CompanyID != authEmployer.CompanyID {
		ctx.JSON(http.StatusUnauthorized, errorResponse(jobOwnershipError))
		return
	}

	err = server.store.DeleteJob(ctx, request.ID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	ctx.JSON(http.StatusNoContent, nil)
}

type updateJobUriRequest struct {
	ID int32 `uri:"id" binding:"required,min=1"`
}

type updateJobRequest struct {
	Title                    string   `json:"title"`
	Description              string   `json:"description"`
	Industry                 string   `json:"industry"`
	Location                 string   `json:"location"`
	SalaryMin                int32    `json:"salary_min"`
	SalaryMax                int32    `json:"salary_max"`
	Requirements             string   `json:"requirements"`
	RequiredSkillsToAdd      []string `json:"required_skills_to_add"`
	RequiredSkillIDsToRemove []int32  `json:"required_skill_ids_to_remove"`
}

// updateJob handles updating a job posting - job and job skills
func (server *Server) updateJob(ctx *gin.Context) {
	// job ID
	var uriRequest updateJobUriRequest
	if err := ctx.ShouldBindUri(&uriRequest); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	// job details
	var request updateJobRequest
	err := json.NewDecoder(ctx.Request.Body).Decode(&request)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	// get employer that is making the request
	authPayload := ctx.MustGet(authorizationPayloadKey).(*token.Payload)
	authEmployer, err := server.store.GetEmployerByEmail(ctx, authPayload.Email)
	if err != nil {
		if err == sql.ErrNoRows {
			ctx.JSON(http.StatusUnauthorized, errorResponse(onlyEmployersAccessError))
			return
		}
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	// get job that is being updated
	job, err := server.store.GetJob(ctx, uriRequest.ID)
	if err != nil {
		if err == sql.ErrNoRows {
			ctx.JSON(http.StatusNotFound, errorResponse(err))
			return
		}

		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	// check if job is owned by the employer
	if job.CompanyID != authEmployer.CompanyID {
		ctx.JSON(http.StatusUnauthorized, errorResponse(jobOwnershipError))
		return
	}

	// update job
	params := db.UpdateJobParams{
		ID:           job.ID,
		Title:        request.Title,
		Description:  request.Description,
		Industry:     request.Industry,
		Location:     request.Location,
		SalaryMin:    request.SalaryMin,
		SalaryMax:    request.SalaryMax,
		Requirements: request.Requirements,
		CompanyID:    job.CompanyID,
	}

	if params.SalaryMin > params.SalaryMax {
		err = fmt.Errorf("salary min cannot be greater than ")
		ctx.JSON(http.StatusBadRequest, errorResponse(salaryRangeError))
		return
	}

	if request.Title == "" {
		params.Title = job.Title
	}
	if request.Description == "" {
		params.Description = job.Description
	}
	if request.Industry == "" {
		params.Industry = job.Industry
	}
	if request.Location == "" {
		params.Location = job.Location
	}
	if request.SalaryMin == 0 {
		params.SalaryMin = job.SalaryMin
	}
	if request.SalaryMax == 0 {
		params.SalaryMax = job.SalaryMax
	}
	if request.Requirements == "" {
		params.Requirements = job.Requirements
	}

	job, err = server.store.UpdateJob(ctx, params)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	// --- update job skills

	// delete
	if len(request.RequiredSkillIDsToRemove) > 0 {
		err = server.store.DeleteMultipleJobSkills(ctx, request.RequiredSkillIDsToRemove)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, errorResponse(err))
			return
		}
	}

	// add
	if len(request.RequiredSkillsToAdd) > 0 {
		err = server.store.CreateMultipleJobSkills(ctx, request.RequiredSkillsToAdd, job.ID)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, errorResponse(err))
			return
		}
	}

	// get skills
	jobSkillsParams := db.ListJobSkillsByJobIDParams{
		JobID:  job.ID,
		Limit:  10,
		Offset: 0,
	}
	jobSkills, err := server.store.ListJobSkillsByJobID(ctx, jobSkillsParams)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	ctx.JSON(http.StatusOK, newJobResponse(job, jobSkills))
}

type getJobRequest struct {
	ID int32 `uri:"id" binding:"required,min=1"`
}

// getJob handles getting a job posting with all details
// without job skills - these are fetched separately
// to allow for the client to get paginated job skills.
func (server *Server) getJob(ctx *gin.Context) {
	var request getJobRequest
	if err := ctx.ShouldBindUri(&request); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	job, err := server.store.GetJobDetails(ctx, request.ID)
	if err != nil {
		if err == sql.ErrNoRows {
			ctx.JSON(http.StatusNotFound, errorResponse(err))
			return
		}
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	ctx.JSON(http.StatusOK, job)
}

type filterAndListJobs struct {
	Title       string `form:"title"`
	Industry    string `form:"industry"`
	JobLocation string `form:"job_location"`
	SalaryMin   int32  `form:"salary_min"`
	SalaryMax   int32  `form:"salary_max"`
	Page        int32  `form:"page" binding:"required,min=1"`
	PageSize    int32  `form:"page_size" binding:"required,min=5,max=15"`
}

func (server *Server) filterAndListJobs(ctx *gin.Context) {
	var request filterAndListJobs
	if err := ctx.ShouldBindQuery(&request); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	params := db.ListJobsByFiltersParams{
		Limit:  request.PageSize,
		Offset: (request.Page - 1) * request.PageSize,
		Title: sql.NullString{
			String: request.Title,
			Valid:  request.Title != "",
		},
		JobLocation: sql.NullString{
			String: request.JobLocation,
			Valid:  request.JobLocation != "",
		},
		Industry: sql.NullString{
			String: request.Industry,
			Valid:  request.Industry != "",
		},
		SalaryMin: sql.NullInt32{
			Int32: request.SalaryMin,
			Valid: request.SalaryMin != 0,
		},
		SalaryMax: sql.NullInt32{
			Int32: request.SalaryMax,
			Valid: request.SalaryMax != 0,
		},
	}

	jobs, err := server.store.ListJobsByFilters(ctx, params)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	ctx.JSON(http.StatusOK, jobs)
}

type listJobsByMatchingSkillsRequest struct {
	Page     int32 `form:"page" binding:"required,min=1"`
	PageSize int32 `form:"page_size" binding:"required,min=5,max=15"`
}

// listJobsByMatchingSkills handles listing all jobs
// that skills match the users skills.
func (server *Server) listJobsByMatchingSkills(ctx *gin.Context) {
	var request listJobsByMatchingSkillsRequest
	if err := ctx.ShouldBindQuery(&request); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	authPayload := ctx.MustGet(authorizationPayloadKey).(*token.Payload)
	authUser, err := server.store.GetUserByEmail(ctx, authPayload.Email)
	if err != nil {
		// person is authenticated but cannot be find in users table
		// means that this is an employer
		if err == sql.ErrNoRows {
			ctx.JSON(http.StatusUnauthorized, errorResponse(onlyUsersAccessError))
		}

		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	params := db.ListJobsMatchingUserSkillsParams{
		UserID: authUser.ID,
		Limit:  request.PageSize,
		Offset: (request.Page - 1) * request.PageSize,
	}

	jobs, err := server.store.ListJobsMatchingUserSkills(ctx, params)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	ctx.JSON(http.StatusOK, jobs)
}
