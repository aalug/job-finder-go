package esearch

import (
	"context"
	"github.com/elastic/go-elasticsearch/v8"
)

// ConnectWithElasticsearch creates a new elasticsearch client and stores it in the context
func ConnectWithElasticsearch(ctx context.Context, address string) context.Context {

	newClient, err := elasticsearch.NewClient(elasticsearch.Config{
		Addresses: []string{
			address,
		},
	})
	if err != nil {
		panic(err)
	}

	return context.WithValue(ctx, ClientKey, newClient)
}
