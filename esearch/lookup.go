package esearch

import (
	"context"
	"encoding/json"
	"strconv"

	"github.com/elastic/go-elasticsearch/v8"
)

// QueryJobsByDocumentID queries the jobs index by document ID.
// This is a helper function for testing purposes.
func QueryJobsByDocumentID(ctx context.Context, documentID int) *Job {

	client := ctx.Value(ClientKey).(*elasticsearch.Client)

	response, err := client.Get("jobs", strconv.Itoa(documentID))
	if err != nil {
		panic(err)
	}
	defer response.Body.Close()

	var getResponse = GetResponse{}
	err = json.NewDecoder(response.Body).Decode(&getResponse)
	if err != nil {
		panic(err)
	}

	return getResponse.Source
}
