package api

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	db "github.com/aalug/go-gin-job-search/internal/db/sqlc"
	"github.com/aalug/go-gin-job-search/internal/worker"
	"github.com/aalug/go-gin-job-search/pkg/token"
	"github.com/aalug/go-gin-job-search/pkg/utils"
	"github.com/aalug/go-gin-job-search/pkg/validation"
	"github.com/gin-gonic/gin"
	"github.com/hibiken/asynq"
	"github.com/lib/pq"
	"net/http"
	"time"
)

var (
	emailNotVerifiedErr = errors.New("email not verified. Please verify your email before logging in")
)

type Skill struct {
	ID                int32  `json:"id"`
	SkillName         string `json:"skill"`
	YearsOfExperience int32  `json:"years_of_experience"`
}

type createUserRequest struct {
	Email             string  `json:"email" binding:"required,email"`
	Password          string  `json:"password" binding:"required,min=6"`
	FullName          string  `json:"full_name" binding:"required"`
	Location          string  `json:"location" binding:"required"`
	DesiredJobTitle   string  `json:"desired_job_title" binding:"required"`
	DesiredIndustry   string  `json:"desired_industry" binding:"required"`
	DesiredSalaryMin  int32   `json:"desired_salary_min" binding:"required,min=0"`
	DesiredSalaryMax  int32   `json:"desired_salary_max" binding:"required,min=0"`
	SkillsDescription string  `json:"skills_description"`
	Experience        string  `json:"experience"`
	Skills            []Skill `json:"skills"`
}

type userResponse struct {
	Email             string    `json:"email"`
	FullName          string    `json:"full_name"`
	Location          string    `json:"location"`
	DesiredJobTitle   string    `json:"desired_job_title"`
	DesiredIndustry   string    `json:"desired_industry"`
	DesiredSalaryMin  int32     `json:"desired_salary_min"`
	DesiredSalaryMax  int32     `json:"desired_salary_max"`
	SkillsDescription string    `json:"skills_description"`
	Experience        string    `json:"experience"`
	Skills            []Skill   `json:"skills"`
	CreatedAt         time.Time `json:"created_at"`
}

// newUserResponse converts db.User to userResponse
func newUserResponse(user db.User, skills []db.UserSkill) userResponse {
	var userSkills []Skill
	for _, skill := range skills {
		userSkills = append(userSkills, Skill{
			ID:                skill.ID,
			SkillName:         skill.Skill,
			YearsOfExperience: skill.Experience,
		})
	}

	return userResponse{
		Email:             user.Email,
		FullName:          user.FullName,
		Location:          user.Location,
		DesiredJobTitle:   user.DesiredJobTitle,
		DesiredIndustry:   user.DesiredIndustry,
		DesiredSalaryMin:  user.DesiredSalaryMin,
		DesiredSalaryMax:  user.DesiredSalaryMax,
		SkillsDescription: user.Skills,
		Experience:        user.Experience,
		Skills:            userSkills,
		CreatedAt:         user.CreatedAt,
	}
}

// @Schemes
// @Summary Create user
// @Description Create a new user
// @Tags users
// @Accept json
// @Produce json
// @param CreateUserRequest body createUserRequest true "User details"
// @Success 201 {object} userResponse
// @Failure 400 {object} ErrorResponse "Invalid request body"
// @Failure 403 {object} ErrorResponse "User with given email already exists"
// @Failure 500 {object} ErrorResponse "Any other error"
// @Router /users [post]
// createUser creates a new user
func (server *Server) createUser(ctx *gin.Context) {
	var request createUserRequest
	if err := ctx.ShouldBindJSON(&request); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	// check if the salary min is not greater than salary max
	if request.DesiredSalaryMin > request.DesiredSalaryMax {
		err := fmt.Errorf("desired salary min is greater than desired salary max")
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	hashedPassword, err := utils.HashPassword(request.Password)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	params := db.CreateUserTxParams{
		CreateUserParams: db.CreateUserParams{
			FullName:         request.FullName,
			Email:            request.Email,
			HashedPassword:   hashedPassword,
			Location:         request.Location,
			DesiredJobTitle:  request.DesiredJobTitle,
			DesiredIndustry:  request.DesiredIndustry,
			DesiredSalaryMin: request.DesiredSalaryMin,
			DesiredSalaryMax: request.DesiredSalaryMax,
			Skills:           request.SkillsDescription,
			Experience:       request.Experience,
		},
		AfterCreate: func(user db.User) error {
			taskPayload := &worker.PayloadSendVerificationEmail{
				Email: user.Email,
			}

			opts := []asynq.Option{
				asynq.MaxRetry(10),
				asynq.ProcessIn(10 * time.Second),
				asynq.Queue(worker.QueueCritical),
			}

			return server.taskDistributor.DistributeTaskSendVerificationEmail(ctx, taskPayload, opts...)
		},
	}

	// Create user
	txResult, err := server.store.CreateUserTx(ctx, params)
	if err != nil {
		if pqErr, ok := err.(*pq.Error); ok {
			switch pqErr.Code.Name() {
			case "unique_violation":
				err := fmt.Errorf("user with this email already exists")
				ctx.JSON(http.StatusForbidden, errorResponse(err))
				return
			}
		}

		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	// Create user skills
	var userSkills []db.UserSkill
	if len(request.Skills) > 0 {
		var skillsParams []db.CreateMultipleUserSkillsParams
		for _, skill := range request.Skills {
			skillsParams = append(skillsParams, db.CreateMultipleUserSkillsParams{
				Skill:      skill.SkillName,
				Experience: skill.YearsOfExperience,
			})
		}

		userSkills, err = server.store.CreateMultipleUserSkills(ctx, skillsParams, txResult.User.ID)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, errorResponse(err))
			return
		}

	}

	res := newUserResponse(txResult.User, userSkills)

	ctx.JSON(http.StatusCreated, res)
}

type loginUserRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6"`
}

type loginUserResponse struct {
	AccessToken string       `json:"access_token"`
	User        userResponse `json:"user"`
}

// @Schemes
// @Summary Login user
// @Description Login user
// @Tags users
// @Accept json
// @Produce json
// @param LoginUserRequest body loginUserRequest true "User credentials"
// @Success 200 {object} loginUserResponse
// @Failure 400 {object} ErrorResponse "Invalid request body"
// @Failure 401 {object} ErrorResponse "Incorrect password"
// @Failure 403 {object} ErrorResponse "Email not verified"
// @Failure 404 {object} ErrorResponse "User with given email does not exist"
// @Failure 500 {object} ErrorResponse "Any other error"
// @Router /users/login [post]
// loginUser handles user login
func (server *Server) loginUser(ctx *gin.Context) {
	var request loginUserRequest
	if err := ctx.ShouldBindJSON(&request); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	user, err := server.store.GetUserByEmail(ctx, request.Email)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			err = fmt.Errorf("user with this email does not exist")
			ctx.JSON(http.StatusNotFound, errorResponse(err))
			return
		}

		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	// check if the user has verified email
	if !user.IsEmailVerified {
		ctx.JSON(http.StatusForbidden, errorResponse(emailNotVerifiedErr))
		return
	}

	err = utils.CheckPassword(request.Password, user.HashedPassword)
	if err != nil {
		err = fmt.Errorf("incorrect password")
		ctx.JSON(http.StatusUnauthorized, errorResponse(err))
		return
	}

	accessToken, err := server.tokenMaker.CreateToken(user.Email, server.config.AccessTokenDuration)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	// get user skills
	params := db.ListUserSkillsParams{
		UserID: user.ID,
		Limit:  10,
		Offset: 0,
	}
	userSkills, err := server.store.ListUserSkills(ctx, params)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	res := loginUserResponse{
		AccessToken: accessToken,
		User:        newUserResponse(user, userSkills),
	}

	ctx.JSON(http.StatusOK, res)
}

// @Schemes
// @Summary Get user
// @Description Get details of the logged-in user
// @Tags users
// @Produce json
// @Success 200 {object} userResponse
// @Failure 401 {object} ErrorResponse "Only users can access this endpoint, not employers."
// @Failure 500 {object} ErrorResponse "Any other error"
// @Security ApiKeyAuth
// @Router /users [get]
// getUser handles getting user details
func (server *Server) getUser(ctx *gin.Context) {
	authPayload := ctx.MustGet(authorizationPayloadKey).(*token.Payload)
	user, userSkills, err := server.store.GetUserDetailsByEmail(ctx, authPayload.Email)
	if err != nil {
		if err == sql.ErrNoRows {
			// but middleware did not stop the request, so we assume
			// that the request was made by an employer
			ctx.JSON(http.StatusUnauthorized, errorResponse(onlyUsersAccessError))
			return
		}

		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	ctx.JSON(http.StatusOK, newUserResponse(user, userSkills))
}

type updateUserRequest struct {
	Email             string  `json:"email"`
	FullName          string  `json:"full_name"`
	Location          string  `json:"location"`
	DesiredJobTitle   string  `json:"desired_job_title"`
	DesiredIndustry   string  `json:"desired_industry"`
	DesiredSalaryMin  int32   `json:"desired_salary_min"`
	DesiredSalaryMax  int32   `json:"desired_salary_max"`
	SkillsDescription string  `json:"skills_description"`
	Experience        string  `json:"experience"`
	SkillsToAdd       []Skill `json:"skills_to_add"`
	SkillsToRemove    []int32 `json:"skill_ids_to_remove"`
}

// @Schemes
// @Summary Update user
// @Description Update the logged-in user details
// @Tags users
// @Accept json
// @Produce json
// @param UpdateUserRequest body updateUserRequest true "User details to update"
// @Success 200 {object} userResponse
// @Failure 400 {object} ErrorResponse "Invalid request body"
// @Failure 401 {object} ErrorResponse "Only users can update their details using this endpoint."
// @Failure 500 {object} ErrorResponse "Any other error"
// @Security ApiKeyAuth
// @Router /users [patch]
// updateUser handles user update
func (server *Server) updateUser(ctx *gin.Context) {
	var request updateUserRequest
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
	authUser, err := server.store.GetUserByEmail(ctx, authPayload.Email)
	if err != nil {
		if err == sql.ErrNoRows {
			// but middleware did not stop the request, so we assume
			// that the request was made by an employer
			ctx.JSON(http.StatusUnauthorized, errorResponse(onlyUsersAccessError))
			return
		}

		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	// check if the salary min is not greater than salary max
	salaryMin := authUser.DesiredSalaryMin
	if request.DesiredSalaryMin != 0 {
		salaryMin = request.DesiredSalaryMin
	}
	salaryMax := authUser.DesiredSalaryMax
	if request.DesiredSalaryMax != 0 {
		salaryMax = request.DesiredSalaryMax
	}

	if salaryMin > salaryMax {
		err := fmt.Errorf("desired salary min is greater than desired salary max")
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	params := db.UpdateUserParams{
		ID:               authUser.ID,
		FullName:         request.FullName,
		Email:            request.Email,
		Location:         request.Location,
		DesiredJobTitle:  request.DesiredJobTitle,
		DesiredIndustry:  request.DesiredIndustry,
		DesiredSalaryMin: salaryMin,
		DesiredSalaryMax: salaryMax,
		Skills:           request.SkillsDescription,
		Experience:       request.Experience,
	}

	if request.Email == "" {
		params.Email = authUser.Email
	}
	if request.FullName == "" {
		params.FullName = authUser.FullName
	}
	if request.Location == "" {
		params.Location = authUser.Location
	}
	if request.DesiredJobTitle == "" {
		params.DesiredJobTitle = authUser.DesiredJobTitle
	}
	if request.DesiredIndustry == "" {
		params.DesiredIndustry = authUser.DesiredIndustry
	}
	if request.SkillsDescription == "" {
		params.Skills = authUser.Skills
	}
	if request.Experience == "" {
		params.Experience = authUser.Experience
	}

	// Update user
	updatedUser, err := server.store.UpdateUser(ctx, params)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	if len(request.SkillsToAdd) > 0 {
		var params []db.CreateMultipleUserSkillsParams
		for _, skill := range request.SkillsToAdd {
			prm := db.CreateMultipleUserSkillsParams{
				Skill:      skill.SkillName,
				Experience: skill.YearsOfExperience,
			}
			params = append(params, prm)
		}

		_, err := server.store.CreateMultipleUserSkills(ctx, params, updatedUser.ID)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, errorResponse(err))
			return
		}
	}

	if len(request.SkillsToRemove) > 0 {
		err = server.store.DeleteMultipleUserSkills(ctx, request.SkillsToRemove)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, errorResponse(err))
			return
		}
	}

	// get all user skills after update
	userSkills, err := server.store.ListUserSkills(ctx, db.ListUserSkillsParams{
		UserID: authUser.ID,
		Limit:  10,
		Offset: 0,
	})
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	ctx.JSON(http.StatusOK, newUserResponse(updatedUser, userSkills))
}

type updateUserPasswordRequest struct {
	OldPassword string `json:"old_password" binding:"required,min=6"`
	NewPassword string `json:"new_password" binding:"required,min=6"`
}

type updateUserPasswordResponse struct {
	Message string `json:"message"`
}

// @Schemes
// @Summary Update user password
// @Description Change / update password of the logged-in user
// @Tags users
// @Accept json
// @Produce json
// @param UpdateUserPasswordRequest body updateUserPasswordRequest true "Users old and new password"
// @Success 200 {object} updateUserPasswordResponse
// @Failure 400 {object} ErrorResponse "Invalid request body"
// @Failure 401 {object} ErrorResponse "Incorrect password or the account making the request is not a user."
// @Failure 500 {object} ErrorResponse "Any other error"
// @Security ApiKeyAuth
// @Router /users/password [patch]
// updateUserPassword handles user password update
func (server *Server) updateUserPassword(ctx *gin.Context) {
	var request updateUserPasswordRequest
	if err := ctx.ShouldBindJSON(&request); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	authPayload := ctx.MustGet(authorizationPayloadKey).(*token.Payload)
	authUser, err := server.store.GetUserByEmail(ctx, authPayload.Email)
	if err != nil {
		if err == sql.ErrNoRows {
			// but middleware did not stop the request, so we assume
			// that the request was made by an employer
			ctx.JSON(http.StatusUnauthorized, errorResponse(onlyUsersAccessError))
			return
		}

		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	err = utils.CheckPassword(request.OldPassword, authUser.HashedPassword)
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

	params := db.UpdatePasswordParams{
		ID:             authUser.ID,
		HashedPassword: hashedPassword,
	}

	err = server.store.UpdatePassword(ctx, params)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	ctx.JSON(http.StatusOK, updateUserPasswordResponse{Message: "password updated successfully"})
}

// @Schemes
// @Summary Delete user
// @Description Delete the logged-in user
// @Tags users
// @Success 204 {null} null
// @Failure 401 {object} ErrorResponse "Only users can update their details using this endpoint."
// @Failure 500 {object} ErrorResponse "Any other error"
// @Security ApiKeyAuth
// @Router /users [delete]
// deleteUser handles deleting users
func (server *Server) deleteUser(ctx *gin.Context) {
	authPayload := ctx.MustGet(authorizationPayloadKey).(*token.Payload)
	authUser, err := server.store.GetUserByEmail(ctx, authPayload.Email)
	if err != nil {
		if err == sql.ErrNoRows {
			// but middleware did not stop the request, so we assume
			// that the request was made by an employer
			ctx.JSON(http.StatusUnauthorized, errorResponse(onlyUsersAccessError))
			return
		}
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	// delete all user skills
	err = server.store.DeleteAllUserSkills(ctx, authUser.ID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	// delete the user
	err = server.store.DeleteUser(ctx, authUser.ID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	ctx.JSON(http.StatusNoContent, nil)
}

type verifyUserEmailRequest struct {
	ID         int64  `form:"id" binding:"required,min=1"`
	SecretCode string `form:"code" binding:"required,min=32"`
}

type verifyUserEmailResponse struct {
	Message string `json:"message"`
}

// @Schemes
// @Summary Verify user email
// @Description Verify user email by providing verify email ID and secret code that should be sent to the user in the verification email.
// @Tags users
// @Param VerifyUserEmailRequest query verifyUserEmailRequest true "Verify user email request"
// @Produce json
// @Success 200 {object} verifyUserEmailResponse
// @Failure 400 {object} ErrorResponse "Invalid request body."
// @Failure 500 {object} ErrorResponse "Any other error."
// @Router /users/verify-email [get]
// verifyUserEmail handles user email verification
func (server *Server) verifyUserEmail(ctx *gin.Context) {
	var request verifyUserEmailRequest
	if err := ctx.ShouldBindQuery(&request); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	txResult, err := server.store.VerifyUserEmailTx(ctx, db.VerifyEmailTxParams{
		ID:         request.ID,
		SecretCode: request.SecretCode,
	})
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	if txResult.User.IsEmailVerified {
		ctx.JSON(http.StatusOK, verifyUserEmailResponse{Message: "Successfully verified email"})
	}
}
