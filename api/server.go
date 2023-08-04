package api

import (
	"fmt"
	db "github.com/aalug/go-gin-job-search/db/sqlc"
	"github.com/aalug/go-gin-job-search/docs"
	"github.com/aalug/go-gin-job-search/esearch"
	"github.com/aalug/go-gin-job-search/token"
	"github.com/aalug/go-gin-job-search/utils"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	swaggerfiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// Server serves HTTP  requests for the service
type Server struct {
	config     utils.Config
	store      db.Store
	tokenMaker token.Maker
	router     *gin.Engine
	esDetails  elasticSearchDetails
}

type elasticSearchDetails struct {
	client            esearch.ESearchClient
	jobs              []esearch.Job
	lastDocumentIndex int64
}

// NewServer creates a new HTTP server and setups routing
func NewServer(config utils.Config, store db.Store, client esearch.ESearchClient) (*Server, error) {
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
		config:     config,
		store:      store,
		tokenMaker: tokenMaker,
		esDetails:  esDetails,
	}

	server.setupRouter()

	return server, nil
}

// setupRouter sets up the HTTP routing
func (server *Server) setupRouter() {
	router := gin.Default()

	routerV1 := router.Group("/api/v1")

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

	// === employers ===
	routerV1.POST("/employers", server.createEmployer)
	routerV1.POST("/employers/login", server.loginEmployer)

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

	// === jobs ===
	// for employers, jobs CRUD
	authRoutesV1.POST("/jobs", server.createJob)
	authRoutesV1.DELETE("/jobs/:id", server.deleteJob)
	authRoutesV1.PATCH("/jobs/:id", server.updateJob)

	// for users, listing jobs that use user details
	authRoutesV1.GET("/jobs/match-skills", server.listJobsByMatchingSkills)

	// === job applications ===
	authRoutesV1.POST("/job-applications", server.createJobApplication)
	authRoutesV1.GET("/job-applications/user/:id", server.getJobApplicationForUser)

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
