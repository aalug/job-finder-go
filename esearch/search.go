package esearch

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/elastic/go-elasticsearch/v8"
	"github.com/elastic/go-elasticsearch/v8/esutil"
)

type ESearchClient interface {
	SearchJobs(ctx context.Context, query string, page, pageSize int32) ([]*Job, error)
	GetDocumentIDByJobID(jobID int) (string, error)
	IndexJobAsDocument(documentID int, job Job) error
	IndexJobsAsDocuments(ctx context.Context) error
	UpdateJobDocument(documentID string, updatedJob Job) error
	DeleteJobDocument(documentID string) error
	QueryJobsByDocumentID(documentID int) *Job
}

type ESClient struct {
	client *elasticsearch.Client
}

func NewClient(client *elasticsearch.Client) ESearchClient {
	return &ESClient{
		client: client,
	}
}

// SearchJobs searches for jobs in the jobs index
func (client ESClient) SearchJobs(ctx context.Context, query string, page, pageSize int32) ([]*Job, error) {
	var jobs []*Job

	var searchBuffer bytes.Buffer
	search := map[string]interface{}{
		"from": (page - 1) * pageSize,
		"size": pageSize,
		"query": map[string]interface{}{
			"bool": map[string]interface{}{
				"should": []interface{}{
					map[string]interface{}{
						"match": map[string]interface{}{
							"title": map[string]interface{}{
								"query":     query,
								"fuzziness": "AUTO",
							},
						},
					},
					map[string]interface{}{
						"multi_match": map[string]interface{}{
							"query": query,
							"fields": []string{
								"description",
								"requirements",
								"job_skills",
								"location",
							},
							"fuzziness": "AUTO",
						},
					},
				},
			},
		},
	}
	err := json.NewEncoder(&searchBuffer).Encode(search)
	if err != nil {
		return jobs, err
	}

	response, err := client.client.Search(
		client.client.Search.WithContext(ctx),
		client.client.Search.WithIndex("jobs"),
		client.client.Search.WithBody(&searchBuffer),
		client.client.Search.WithTrackTotalHits(true),
		client.client.Search.WithPretty(),
	)
	if err != nil {
		return jobs, err
	}
	defer response.Body.Close()

	var searchResponse = SearchResponse{}
	err = json.NewDecoder(response.Body).Decode(&searchResponse)
	if err != nil {
		return jobs, err
	}

	if searchResponse.Hits.Total.Value > 0 {
		for _, job := range searchResponse.Hits.Hits {
			jobs = append(jobs, job.Source)
		}
	}
	return jobs, nil
}

// GetDocumentIDByJobID gets the document ID of a job by job ID.
func (client ESClient) GetDocumentIDByJobID(jobID int) (string, error) {
	// Create a Term Query to match the job ID field with the provided jobID.
	query := map[string]interface{}{
		"query": map[string]interface{}{
			"term": map[string]interface{}{
				"id": jobID,
			},
		},
	}

	// Perform the search request with the Term Query.
	response, err := client.client.Search(
		client.client.Search.WithIndex("jobs"),
		client.client.Search.WithBody(esutil.NewJSONReader(query)),
	)
	if err != nil {
		return "", err
	}
	defer response.Body.Close()

	// Parse the search response to get the documentID.
	var searchResponse = SearchResponse{}
	err = json.NewDecoder(response.Body).Decode(&searchResponse)
	if err != nil {
		return "", err
	}

	// Extract the documentID from the search response.
	if searchResponse.Hits.Total.Value > 0 {
		documentID := searchResponse.Hits.Hits[0].ID
		return documentID, nil
	}

	// Return an error if no document is found with the given jobID.
	return "", fmt.Errorf("no document found with jobID: %d", jobID)
}
