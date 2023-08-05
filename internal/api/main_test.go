package api

import (
	"github.com/aalug/go-gin-job-search/internal/config"
	"github.com/aalug/go-gin-job-search/internal/db/sqlc"
	"github.com/aalug/go-gin-job-search/internal/esearch"
	"github.com/aalug/go-gin-job-search/pkg/utils"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
	"os"
	"testing"
	"time"
)

func newTestServer(t *testing.T, store db.Store, client esearch.ESearchClient) *Server {
	cfg := config.Config{
		TokenSymmetricKey:   utils.RandomString(32),
		AccessTokenDuration: time.Minute,
	}

	server, err := NewServer(cfg, store, client)
	require.NoError(t, err)

	if client != nil {
		server.esDetails.lastDocumentIndex = 1
	}

	return server
}

func TestMain(m *testing.M) {
	gin.SetMode(gin.TestMode)

	os.Exit(m.Run())
}
