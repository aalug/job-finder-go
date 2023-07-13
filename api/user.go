package api

import (
	"database/sql"
	"encoding/json"
	"fmt"
	db "github.com/aalug/go-gin-job-search/db/sqlc"
	"github.com/aalug/go-gin-job-search/token"
	"github.com/aalug/go-gin-job-search/utils"
	"github.com/aalug/go-gin-job-search/validation"
	"github.com/gin-gonic/gin"
	"github.com/lib/pq"
	"net/http"
	"time"
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

	params := db.CreateUserParams{
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
	}

	// Create user
	user, err := server.store.CreateUser(ctx, params)
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

		userSkills, err = server.store.CreateMultipleUserSkills(ctx, skillsParams, user.ID)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, errorResponse(err))
			return
		}

	}

	res := newUserResponse(user, userSkills)

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

// loginUser handles user login
func (server *Server) loginUser(ctx *gin.Context) {
	var request loginUserRequest
	if err := ctx.ShouldBindJSON(&request); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	user, err := server.store.GetUserByEmail(ctx, request.Email)
	if err != nil {
		if err == sql.ErrNoRows {
			err = fmt.Errorf("user with this email does not exist")
			ctx.JSON(http.StatusNotFound, errorResponse(err))
			return
		}
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
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

// getUser handles getting user details
func (server *Server) getUser(ctx *gin.Context) {
	authPayload := ctx.MustGet(authorizationPayloadKey).(*token.Payload)
	user, userSkills, err := server.store.GetUserDetailsByEmail(ctx, authPayload.Email)
	if err != nil {
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

	ctx.JSON(http.StatusOK, gin.H{"message": "password updated successfully"})
}

// deleteUser handles deleting users
func (server *Server) deleteUser(ctx *gin.Context) {
	authPayload := ctx.MustGet(authorizationPayloadKey).(*token.Payload)
	authUser, err := server.store.GetUserByEmail(ctx, authPayload.Email)
	if err != nil {
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
