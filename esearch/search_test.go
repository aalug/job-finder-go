package esearch

import (
	"context"
	"github.com/elastic/go-elasticsearch/v8"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"strconv"
	"testing"
)

func TestSearchJobs(t *testing.T) {
	ctx := context.Background()
	c, err := elasticsearch.NewDefaultClient()
	require.NoError(t, err)
	client := ESClient{client: c}

	jobID := 1
	jobs := []Job{
		{
			ID:          int32(jobID),
			Title:       "Software Engineer",
			Description: "Job description...",
			Location:    "New York",
		},
		{
			ID:          2,
			Title:       "Data Scientist",
			Description: "Data Scientist description...",
			Location:    "San Francisco",
		},
	}
	ctx = context.WithValue(ctx, JobKey, jobs)
	err = client.IndexJobsAsDocuments(ctx)
	require.NoError(t, err)

	results, err := client.SearchJobs(ctx, "software engineer", 1, 10)
	require.NoError(t, err)

	jobsFromContext, ok := ctx.Value(JobKey).([]Job)
	require.True(t, ok)

	require.Equal(t, jobsFromContext[0].ID, results[0].ID)
	require.Equal(t, jobsFromContext[0].Title, results[0].Title)
	require.Equal(t, jobsFromContext[0].Description, results[0].Description)
	require.Equal(t, jobsFromContext[0].Location, results[0].Location)

	// GetDocumentIDByJobID tests
	documentID, err := client.GetDocumentIDByJobID(jobID)
	require.NoError(t, err)
	require.NoError(t, err)
	require.Equal(t, documentID, strconv.Itoa(jobID))

	documentID2, err := client.GetDocumentIDByJobID(9999999999)
	require.Error(t, err)
	require.Empty(t, documentID2)
}

func TestNewClient(t *testing.T) {
	c, err := elasticsearch.NewDefaultClient()
	assert.NoError(t, err)

	// Call the function being tested
	client := NewClient(c)

	require.Equal(t, c, client.(*ESClient).client, "ESearchClient should contain the same Elasticsearch client")
}
