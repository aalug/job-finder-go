package api

import (
	"fmt"
	"github.com/aalug/go-gin-job-search/docs"
	"github.com/aalug/go-gin-job-search/internal/config"
	"github.com/aalug/go-gin-job-search/internal/db/sqlc"
	"github.com/aalug/go-gin-job-search/internal/esearch"
	"github.com/aalug/go-gin-job-search/internal/worker"
	"github.com/aalug/go-gin-job-search/pkg/token"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	swaggerfiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

var BaseUrl string

// Server serves HTTP  requests for the service
type Server struct {
	config          config.Config
	store           db.Store
	tokenMaker      token.Maker
	router          *gin.Engine
	esDetails       elasticSearchDetails
	taskDistributor worker.TaskDistributor
}

type elasticSearchDetails struct {
	client            esearch.ESearchClient
	jobs              []esearch.Job
	lastDocumentIndex int64
}

// NewServer creates a new HTTP server and setups routing
func NewServer(config config.Config, store db.Store, client esearch.ESearchClient, taskDistributor worker.TaskDistributor) (*Server, error) {
	BaseUrl = config.BaseUrl

	// === tokens ===
	tokenMaker, err := token.NewPasetoMaker(config.TokenSymmetricKey)
	if err != nil {
		return nil, fmt.Errorf("cannot create token maker: %w", err)
	}

	// === elasticsearch ===
	esDetails := elasticSearchDetails{
		client: client,
	}

	server := &Server{
		config:          config,
		store:           store,
		tokenMaker:      tokenMaker,
		esDetails:       esDetails,
		taskDistributor: taskDistributor,
	}

	server.setupRouter()

	return server, nil
}

// setupRouter sets up the HTTP routing
func (server *Server) setupRouter() {
	router := gin.Default()

	routerV1 := router.Group(BaseUrl)

	// CORS
	corsConfig := cors.DefaultConfig()
	corsConfig.AllowAllOrigins = true
	corsConfig.AllowHeaders = append(corsConfig.AllowHeaders, "Authorization")
	router.Use(cors.New(corsConfig))

	// Swagger docs
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerfiles.Handler))
	docs.SwaggerInfo.BasePath = "/api/v1"

	// === users ===
	routerV1.POST("/users", server.createUser)
	routerV1.POST("/users/login", server.loginUser)
	routerV1.GET("/users/verify-email", server.verifyUserEmail)
	routerV1.GET("/users/send-verification-email", server.sendVerificationEmailToUser)

	// === employers ===
	routerV1.POST("/employers", server.createEmployer)
	routerV1.POST("/employers/login", server.loginEmployer)
	routerV1.GET("/employers/verify-email", server.verifyEmployerEmail)

	routerV1.GET("/employers/employer-company-details/:email", server.getEmployerAndCompanyDetails)

	// === jobs ===
	routerV1.GET("/jobs/:id", server.getJob)
	routerV1.GET("/jobs", server.filterAndListJobs)
	routerV1.GET("/jobs/company", server.listJobsByCompany)
	routerV1.GET("/jobs/search", server.searchJobs)

	// ===== routes that require authentication =====
	authRoutesV1 := routerV1.Group("/").Use(authMiddleware(server.tokenMaker))

	// === users ===
	authRoutesV1.GET("/users", server.getUser)
	authRoutesV1.PATCH("/users", server.updateUser)
	authRoutesV1.PATCH("/users/password", server.updateUserPassword)
	authRoutesV1.DELETE("/users", server.deleteUser)

	// === employers ===
	authRoutesV1.GET("/employers", server.getEmployer)
	authRoutesV1.PATCH("/employers", server.updateEmployer)
	authRoutesV1.PATCH("/employers/password", server.updateEmployerPassword)
	authRoutesV1.DELETE("/employers", server.deleteEmployer)
	authRoutesV1.GET("/employers/user-details/:email", server.getUserAsEmployer)

	// === jobs ===
	// for employers, jobs CRUD
	authRoutesV1.POST("/jobs", server.createJob)
	authRoutesV1.GET("/jobs/employer", server.listEmployerJobs)
	authRoutesV1.PATCH("/jobs/:id", server.updateJob)
	authRoutesV1.DELETE("/jobs/:id", server.deleteJob)

	// for users, listing jobs that use user details
	authRoutesV1.GET("/jobs/match-skills", server.listJobsByMatchingSkills)

	// === job applications ===
	// for users, job applications CRUD
	authRoutesV1.POST("/job-applications", server.createJobApplication)
	authRoutesV1.GET("/job-applications/user/:id", server.getJobApplicationForUser)
	authRoutesV1.PATCH("/job-applications/user/:id", server.updateJobApplication)
	authRoutesV1.DELETE("/job-applications/user/:id", server.deleteJobApplication)
	authRoutesV1.GET("/job-applications/user", server.listJobApplicationsForUser)

	// for employers, reading, changing statuses (rejecting, offering)
	authRoutesV1.GET("/job-applications/employer/:id", server.getJobApplicationForEmployer)
	authRoutesV1.PATCH("/job-applications/employer/:id/status", server.changeJobApplicationStatus)
	authRoutesV1.GET("/job-applications/employer", server.listJobApplicationsForEmployer)

	server.router = router
}

// Start runs the HTTP server on a given address
func (server *Server) Start(address string) error {
	return server.router.Run(address)
}

type ErrorResponse struct {
	Error string `json:"error"`
}

func errorResponse(err error) ErrorResponse {
	return ErrorResponse{Error: err.Error()}
}
