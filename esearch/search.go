package esearch

import (
	"bytes"
	"context"
	"encoding/json"
	"github.com/elastic/go-elasticsearch/v8"
)

type ESearchClient interface {
	SearchJobs(ctx context.Context, query string, page, pageSize int32) ([]*Job, error)
}

type ESClient struct {
	client *elasticsearch.Client
}

func NewClient(client *elasticsearch.Client) ESearchClient {
	return &ESClient{
		client: client,
	}
}

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
