package esearch

import (
	"github.com/elastic/go-elasticsearch/v8"
)

// ConnectWithElasticsearch creates a new elasticsearch client and stores it in the context
func ConnectWithElasticsearch(address string) (*elasticsearch.Client, error) {

	newClient, err := elasticsearch.NewClient(elasticsearch.Config{
		Addresses: []string{
			address,
		},
	})
	if err != nil {
		return nil, err
	}

	//return context.WithValue(ctx, ClientKey, newClient), nil
	return newClient, nil
}
