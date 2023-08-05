package esearch

import (
	"github.com/stretchr/testify/require"
	"os"
	"testing"
)

func TestConnectWithElasticsearch(t *testing.T) {
	esAddress := os.Getenv("ELASTICSEARCH_ADDRESS")

	client, err := ConnectWithElasticsearch(esAddress)
	require.NoError(t, err)

	require.NotNil(t, client, "Returned client should not be nil")
}
