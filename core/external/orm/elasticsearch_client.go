package orm

import (
	"context"
	"log/slog"
	"strings"
	"time"

	elasticsearch_v8 "github.com/elastic/go-elasticsearch/v8"
	elasticsearch_esutil_v8 "github.com/elastic/go-elasticsearch/v8/esutil"
	elasticsearch_query_v8 "github.com/elastic/go-elasticsearch/v8/typedapi/esql/query"
	elasticsearch_v9 "github.com/elastic/go-elasticsearch/v9"
	elasticsearch_esutil_v9 "github.com/elastic/go-elasticsearch/v9/esutil"
	elasticsearch_query_v9 "github.com/elastic/go-elasticsearch/v9/typedapi/esql/query"
	"ntels.com/pharos/core/external"
)

func NewElasticsearchClient(elasticsearchConfig ElasticsearchConfig) (*ElasticsearchClient, error) {
	duration, err := time.ParseDuration(elasticsearchConfig.Bulk.FlushInterval)
	if err != nil {
		return nil, err
	}

	switch elasticsearchConfig.Version {
	case 8:
		cfg := elasticsearch_v8.Config{
			Addresses: elasticsearchConfig.Addresses,
		}

		client, err := elasticsearch_v8.NewTypedClient(cfg)
		if err != nil {
			return nil, err
		}

		bulkIndexerV8, err := elasticsearch_esutil_v8.NewBulkIndexer(elasticsearch_esutil_v8.BulkIndexerConfig{
			FlushBytes:    elasticsearchConfig.Bulk.FlushBytes,
			FlushInterval: duration,
			Client:        client,
		})
		if err != nil {
			return nil, err
		}

		return &ElasticsearchClient{version: 8, v8: client, bulkIndexerV8: bulkIndexerV8}, nil
	case 9:
		client, err := elasticsearch_v9.NewTyped(elasticsearch_v9.WithAddresses(elasticsearchConfig.Addresses...))
		if err != nil {
			return nil, err
		}

		bulkIndexerV9, err := elasticsearch_esutil_v9.NewBulkIndexer(elasticsearch_esutil_v9.BulkIndexerConfig{
			FlushBytes:    elasticsearchConfig.Bulk.FlushBytes,
			FlushInterval: duration,
			Client:        client,
		})
		if err != nil {
			return nil, err
		}

		return &ElasticsearchClient{version: 9, v9: client, bulkIndexerV9: bulkIndexerV9}, nil
	default:
		return nil, external.ErrorNotSupported
	}
}

type ElasticsearchClient struct {
	version int

	v8 *elasticsearch_v8.TypedClient
	v9 *elasticsearch_v9.TypedClient

	bulkIndexerV8 elasticsearch_esutil_v8.BulkIndexer
	bulkIndexerV9 elasticsearch_esutil_v9.BulkIndexer
}

func (ec *ElasticsearchClient) Close(timeout int) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(timeout)*time.Second)
	defer cancel()

	switch ec.version {
	case 8:
		if ec.bulkIndexerV8 != nil {
			if err := ec.bulkIndexerV8.Close(ctx); err != nil {
				slog.Error("failed to close bulk indexer v8", "error", err)
			}
		}
	case 9:
		if ec.bulkIndexerV9 != nil {
			if err := ec.bulkIndexerV9.Close(ctx); err != nil {
				slog.Error("failed to close bulk indexer v9", "error", err)
			}
		}
	default:
		return nil
	}

	return nil
}

func (ec *ElasticsearchClient) GetVersion() int {
	return ec.version
}

func (ec *ElasticsearchClient) IndexBulk(index string, documentID, document string) error {
	switch ec.version {
	case 8:
		if ec.v8 == nil || ec.bulkIndexerV8 == nil {
			return external.ErrorNotConnected
		}

		return ec.indexBulkV8(index, documentID, document)
	case 9:
		if ec.v9 == nil || ec.bulkIndexerV9 == nil {
			return external.ErrorNotConnected
		}

		return ec.indexBulkV9(index, documentID, document)
	default:
		return external.ErrorNotSupported
	}
}

func (ec *ElasticsearchClient) indexBulkV8(index string, documentID, document string) error {
	return ec.bulkIndexerV8.Add(
		context.Background(),

		elasticsearch_esutil_v8.BulkIndexerItem{
			Index:      index,
			Action:     "index",
			DocumentID: documentID,
			Body:       strings.NewReader(document),

			OnSuccess: func(
				ctx context.Context,
				item elasticsearch_esutil_v8.BulkIndexerItem,
				res elasticsearch_esutil_v8.BulkIndexerResponseItem,
			) {
				slog.Debug("Document indexed successfully", "index", item.Index, "document_id", item.DocumentID)
			},
			OnFailure: func(
				ctx context.Context,
				item elasticsearch_esutil_v8.BulkIndexerItem,
				res elasticsearch_esutil_v8.BulkIndexerResponseItem, err error,
			) {
				if err != nil {
					slog.Error("Failed to index document", "error", err)
				} else {
					slog.Error("Failed to index document", "status", res.Status, "error", res.Error)
				}
			},
		},
	)
}

func (ec *ElasticsearchClient) indexBulkV9(index string, documentID, document string) error {
	return ec.bulkIndexerV9.Add(
		context.Background(),

		elasticsearch_esutil_v9.BulkIndexerItem{
			Index:      index,
			Action:     "index",
			DocumentID: documentID,
			Body:       strings.NewReader(document),

			OnSuccess: func(
				ctx context.Context,
				item elasticsearch_esutil_v9.BulkIndexerItem,
				res elasticsearch_esutil_v9.BulkIndexerResponseItem,
			) {
				slog.Debug("Document indexed successfully", "index", item.Index, "document_id", item.DocumentID)
			},
			OnFailure: func(
				ctx context.Context,
				item elasticsearch_esutil_v9.BulkIndexerItem,
				res elasticsearch_esutil_v9.BulkIndexerResponseItem, err error,
			) {
				if err != nil {
					slog.Error("Failed to index document", "error", err)
				} else {
					slog.Error("Failed to index document", "status", res.Status, "error", res.Error)
				}
			},
		},
	)
}

func (ec *ElasticsearchClient) QueryForESQL(ctx context.Context, query string, timeout int) (DatabaseResponse, error) {
	switch ec.version {
	case 8:
		return ec.queryForESQLV8(ctx, query, timeout)
	case 9:
		return ec.queryForESQLV9(ctx, query, timeout)
	default:
		return DatabaseResponse{}, external.ErrorNotSupported
	}
}

func (ec *ElasticsearchClient) queryForESQLV8(ctx context.Context, query string, timeout int) (DatabaseResponse, error) {
	client := ec.v8
	if client == nil {
		return DatabaseResponse{}, external.ErrorNotConnected
	}

	start := time.Now()

	ctx, cancel := context.WithTimeout(ctx, time.Duration(timeout)*time.Second)
	defer cancel()

	datas, err := elasticsearch_query_v8.Helper[map[string]any](ctx, client.Esql.Query().Query(query))
	if err != nil {
		return DatabaseResponse{}, err
	}

	return ec.setDatabaseResponse(query, start, datas)
}

func (ec *ElasticsearchClient) queryForESQLV9(ctx context.Context, query string, timeout int) (DatabaseResponse, error) {
	client := ec.v9
	if client == nil {
		return DatabaseResponse{}, external.ErrorNotConnected
	}

	start := time.Now()

	ctx, cancel := context.WithTimeout(ctx, time.Duration(timeout)*time.Second)
	defer cancel()

	datas, err := elasticsearch_query_v9.Helper[map[string]any](ctx, client.Esql.Query().Query(query))
	if err != nil {
		return DatabaseResponse{}, err
	}

	return ec.setDatabaseResponse(query, start, datas)
}

func (ec *ElasticsearchClient) setDatabaseResponse(query string, start time.Time, datas []map[string]any) (DatabaseResponse, error) {
	var response DatabaseResponse
	response.Statistics.Elapsed = time.Since(start).Seconds()
	response.SQL = query
	response.Rows = int64(len(datas))
	response.Data = make([]map[string]any, 0, len(datas))

	// 첫 번째 행에서 컬럼 정보 추출 (순서 보장)
	if len(datas) > 0 {
		firstRow := datas[0]
		response.Meta = make([]map[string]string, 0, len(firstRow))

		// 컬럼 순서를 유지하기 위해 첫 행의 키 순서대로 메타데이터 생성
		for name := range firstRow {
			response.Meta = append(response.Meta, map[string]string{
				"name": name,
			})
		}
	}

	// 데이터 복사
	response.Data = append(response.Data, datas...)
	return response, nil
}
