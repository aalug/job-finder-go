package esearch

import (
	"bytes"
	"context"
	"encoding/json"
	"github.com/elastic/go-elasticsearch/v8"
)

func SearchJobs(ctx context.Context, query string) []*Job {

	client := ctx.Value(ClientKey).(*elasticsearch.Client)

	var searchBuffer bytes.Buffer
	search := map[string]interface{}{
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
		panic(err)
	}

	response, err := client.Search(
		client.Search.WithContext(ctx),
		client.Search.WithIndex("jobs"),
		client.Search.WithBody(&searchBuffer),
		client.Search.WithTrackTotalHits(true),
		client.Search.WithPretty(),
	)
	if err != nil {
		panic(err)
	}
	defer response.Body.Close()

	var searchResponse = SearchResponse{}
	err = json.NewDecoder(response.Body).Decode(&searchResponse)
	if err != nil {
		panic(err)
	}

	var jobs []*Job
	if searchResponse.Hits.Total.Value > 0 {
		for _, job := range searchResponse.Hits.Hits {
			jobs = append(jobs, job.Source)
		}
	}
	return jobs
}
