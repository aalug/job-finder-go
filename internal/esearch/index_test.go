package esearch

import (
	"bytes"
	"context"
	"github.com/aalug/go-gin-job-search/pkg/utils"
	"github.com/elastic/go-elasticsearch/v8"
	"github.com/elastic/go-elasticsearch/v8/esutil"
	"github.com/stretchr/testify/require"
	"strconv"
	"testing"
)

func TestIndexJobsAsDocuments(t *testing.T) {
	// Create a mock Elasticsearch client
	mockClient, err := elasticsearch.NewDefaultClient()
	require.NoError(t, err)

	client := ESClient{
		client: mockClient,
	}

	var jobs []Job
	for i := 0; i < 10; i++ {
		// Create a random job and add it to the slice
		jobs = append(jobs, createRandomJob())
	}

	// Set up context with the jobs slice
	ctx := context.WithValue(context.Background(), JobKey, jobs)

	// Create a mock bulk indexer
	mockBulkIndexer, err := esutil.NewBulkIndexer(esutil.BulkIndexerConfig{
		Index:      "jobs",
		Client:     client.client,
		NumWorkers: 5,
	})
	require.NoError(t, err)

	for documentID, document := range jobs {
		body, err := readerToReadSeeker(esutil.NewJSONReader(document))
		require.NoError(t, err)
		err = mockBulkIndexer.Add(
			ctx,
			esutil.BulkIndexerItem{
				Action:     "index",
				DocumentID: strconv.Itoa(documentID),
				Body:       body,
			},
		)
		require.NoError(t, err)
	}
	mockBulkIndexer.Close(ctx)

	// Call the function being tested
	err = client.IndexJobsAsDocuments(ctx)
	require.NoError(t, err)
}

func TestIndexJobAsDocument(t *testing.T) {
	// Create a mock Elasticsearch client
	mockClient, err := elasticsearch.NewDefaultClient()
	require.NoError(t, err)

	// Replace the real client with the mock client
	client := ESClient{
		client: mockClient,
	}

	// Test data
	documentID := 123
	job := createRandomJob()
	err = client.IndexJobAsDocument(documentID, job)
	require.NoError(t, err)
}

func TestUpdateJobDocument(t *testing.T) {
	// Create a mock Elasticsearch client
	mockClient, err := elasticsearch.NewDefaultClient()
	require.NoError(t, err)

	// Replace the real client with the mock client
	client := ESClient{
		client: mockClient,
	}

	// Test data
	documentID := "job_123" // Assuming a valid document ID
	updatedJob := createRandomJob()

	err = client.UpdateJobDocument(documentID, updatedJob)
	require.NoError(t, err)
}

func TestDeleteJobDocument(t *testing.T) {
	// Create a mock Elasticsearch client
	mockClient, err := elasticsearch.NewDefaultClient()
	require.NoError(t, err)

	client := ESClient{
		client: mockClient,
	}

	documentID := "job_123" // Assuming a valid document ID

	err = client.DeleteJobDocument(documentID)
	require.NoError(t, err)
}

func TestReaderToReadSeeker(t *testing.T) {
	// Test data
	inputData := "Hello, world!"
	reader := bytes.NewBufferString(inputData)

	readSeeker, err := readerToReadSeeker(reader)

	require.NoError(t, err)
	require.NotNil(t, readSeeker, "Expected a valid io.ReadSeeker")

	// Ensure the content in the io.ReadSeeker matches the input data
	buffer := new(bytes.Buffer)
	_, err = buffer.ReadFrom(readSeeker)
	require.NoError(t, err)
	require.Equal(t, inputData, buffer.String(), "Input data and readSeeker content mismatch")
}

func TestReaderToReadSeekerEmptyReader(t *testing.T) {
	// Test data: Empty reader
	reader := bytes.NewBuffer([]byte{})

	readSeeker, err := readerToReadSeeker(reader)

	require.NoError(t, err)
	require.NotNil(t, readSeeker, "Expected a valid io.ReadSeeker")

	// Ensure the content in the io.ReadSeeker is empty
	buffer := new(bytes.Buffer)
	_, err = buffer.ReadFrom(readSeeker)
	require.NoError(t, err)
	require.Empty(t, buffer.String(), "Expected empty readSeeker content")
}

func createRandomJob() Job {
	return Job{
		ID:           utils.RandomInt(1, 100),
		Title:        utils.RandomString(5),
		Industry:     utils.RandomString(3),
		CompanyName:  utils.RandomString(5),
		Description:  utils.RandomString(5),
		Location:     utils.RandomString(2),
		SalaryMin:    utils.RandomInt(0, 100),
		SalaryMax:    utils.RandomInt(100, 200),
		Requirements: utils.RandomString(6),
		JobSkills:    []string{utils.RandomString(2), utils.RandomString(2)},
	}
}
