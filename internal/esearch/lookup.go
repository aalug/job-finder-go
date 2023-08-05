package esearch

import (
	"encoding/json"
	"strconv"
)

// QueryJobsByDocumentID queries the jobs index by document ID.
// This is a helper function for testing purposes.
func (client ESClient) QueryJobsByDocumentID(documentID int) *Job {
	response, err := client.client.Get("jobs", strconv.Itoa(documentID))
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
