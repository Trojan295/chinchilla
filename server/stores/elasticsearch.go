package stores

import (
	"context"
	"encoding/json"

	"github.com/Trojan295/chinchilla/server/gameservers"
	"github.com/olivere/elastic/v7"
	"github.com/olivere/elastic/v7/config"
)

type ElasticsearchStore struct {
	address string
	client  *elastic.Client
}

func NewElasticsearchStore(address string) ElasticsearchStore {
	sniff := false
	cfg, _ := config.Parse(address)
	cfg.Sniff = &sniff
	client, err := elastic.NewClientFromConfig(cfg)
	if err != nil {
		panic(err)
	}

	return ElasticsearchStore{
		address: address,
		client:  client,
	}
}

func (store *ElasticsearchStore) GetLogs(request *gameservers.GetLogsRequest) (*gameservers.GetLogsResponse, error) {

	termQuery := elastic.NewBoolQuery().
		Must(elastic.NewMatchQuery("container.name", request.GameserverUUID))
	searchResult, err := store.client.Search().
		Index("filebeat-*").
		From(0).
		Size(request.Lines).
		Query(termQuery).
		Do(context.Background())
	if err != nil {
		return nil, err
	}

	logs := make([]string, 0)
	for _, hit := range searchResult.Hits.Hits {
		var x map[string]string
		json.Unmarshal(hit.Source, &x)

		message := x["message"]
		logs = append(logs, message)
	}

	return &gameservers.GetLogsResponse{
		Logs: logs,
	}, nil
}
