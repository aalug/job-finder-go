package api

import (
	db "github.com/aalug/go-gin-job-search/db/sqlc"
	"github.com/aalug/go-gin-job-search/utils"
	"github.com/gin-gonic/gin"
	"os"
	"testing"
)

func newTestServer(t *testing.T, store db.Store) *Server {
	config := utils.Config{}

	server := NewServer(config, store)

	return server
}

func TestMain(m *testing.M) {
	gin.SetMode(gin.TestMode)

	os.Exit(m.Run())
}
