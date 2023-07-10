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

type Skill struct {
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
