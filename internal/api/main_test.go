package api

import (
	"github.com/aalug/job-finder-go/internal/config"
	"github.com/aalug/job-finder-go/internal/db/sqlc"
	"github.com/aalug/job-finder-go/internal/esearch"
	"github.com/aalug/job-finder-go/internal/worker"
	"github.com/aalug/job-finder-go/pkg/utils"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
	"os"
	"testing"
	"time"
)

func newTestServer(t *testing.T, store db.Store, client esearch.ESearchClient, taskDistributor worker.TaskDistributor) *Server {
	cfg := config.Config{
		TokenSymmetricKey:   utils.RandomString(32),
		AccessTokenDuration: time.Minute,
	}

	server, err := NewServer(cfg, store, client, taskDistributor)
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
