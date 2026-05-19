package orm

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"ntels.com/pharos/core/external"
)

func TestNewElasticsearchClient_Version8(t *testing.T) {
	// 실제 Elasticsearch 서버 없이는 연결 실패 예상
	config := ElasticsearchConfig{
		Version:   8,
		Addresses: []string{"http://localhost:9200"},
		Bulk: struct {
			FlushBytes    int    `mapstructure:"flush_bytes"`
			FlushInterval string `mapstructure:"flush_interval"`
		}{
			FlushBytes:    5000000,
			FlushInterval: "30s",
		},
	}

	client, err := NewElasticsearchClient(config)
	// 연결은 성공하지만 클라이언트는 생성되어야 함
	if err == nil {
		require.NotNil(t, client)
		assert.Equal(t, 8, client.GetVersion())
		defer client.Close(5)
	}
}

func TestNewElasticsearchClient_Version9(t *testing.T) {
	config := ElasticsearchConfig{
		Version:   9,
		Addresses: []string{"http://localhost:9200"},
		Bulk: struct {
			FlushBytes    int    `mapstructure:"flush_bytes"`
			FlushInterval string `mapstructure:"flush_interval"`
		}{
			FlushBytes:    5000000,
			FlushInterval: "30s",
		},
	}

	client, err := NewElasticsearchClient(config)
	if err == nil {
		require.NotNil(t, client)
		assert.Equal(t, 9, client.GetVersion())
		defer client.Close(5)
	}
}

func TestNewElasticsearchClient_UnsupportedVersion(t *testing.T) {
	config := ElasticsearchConfig{
		Version:   7, // 지원하지 않는 버전
		Addresses: []string{"http://localhost:9200"},
		Bulk: struct {
			FlushBytes    int    `mapstructure:"flush_bytes"`
			FlushInterval string `mapstructure:"flush_interval"`
		}{
			FlushBytes:    5000000,
			FlushInterval: "30s",
		},
	}

	client, err := NewElasticsearchClient(config)
	assert.Error(t, err)
	assert.Equal(t, external.ErrorNotSupported, err)
	assert.Nil(t, client)
}

func TestNewElasticsearchClient_InvalidFlushInterval(t *testing.T) {
	config := ElasticsearchConfig{
		Version:   8,
		Addresses: []string{"http://localhost:9200"},
		Bulk: struct {
			FlushBytes    int    `mapstructure:"flush_bytes"`
			FlushInterval string `mapstructure:"flush_interval"`
		}{
			FlushBytes:    5000000,
			FlushInterval: "invalid", // 잘못된 duration 형식
		},
	}

	client, err := NewElasticsearchClient(config)
	assert.Error(t, err)
	assert.Nil(t, client)
}

func TestElasticsearchClient_GetVersion(t *testing.T) {
	config8 := ElasticsearchConfig{
		Version:   8,
		Addresses: []string{"http://localhost:9200"},
		Bulk: struct {
			FlushBytes    int    `mapstructure:"flush_bytes"`
			FlushInterval string `mapstructure:"flush_interval"`
		}{
			FlushBytes:    5000000,
			FlushInterval: "30s",
		},
	}

	client, err := NewElasticsearchClient(config8)
	if err == nil {
		defer client.Close(5)
		assert.Equal(t, 8, client.GetVersion())
	}
}

func TestElasticsearchClient_Close(t *testing.T) {
	config := ElasticsearchConfig{
		Version:   8,
		Addresses: []string{"http://localhost:9200"},
		Bulk: struct {
			FlushBytes    int    `mapstructure:"flush_bytes"`
			FlushInterval string `mapstructure:"flush_interval"`
		}{
			FlushBytes:    5000000,
			FlushInterval: "30s",
		},
	}

	client, err := NewElasticsearchClient(config)
	if err == nil {
		err = client.Close(5)
		assert.NoError(t, err)
	}
}

func TestElasticsearchClient_IndexBulk_NotConnected(t *testing.T) {
	// 연결되지 않은 클라이언트
	client := &ElasticsearchClient{
		version: 8,
		v8:      nil,
	}

	err := client.IndexBulk("test-index", "doc-1", `{"field":"value"}`)
	assert.Error(t, err)
	assert.Equal(t, external.ErrorNotConnected, err)
}

func TestElasticsearchClient_IndexBulk_UnsupportedVersion(t *testing.T) {
	client := &ElasticsearchClient{
		version: 7, // 지원하지 않는 버전
	}

	err := client.IndexBulk("test-index", "doc-1", `{"field":"value"}`)
	assert.Error(t, err)
	assert.Equal(t, external.ErrorNotSupported, err)
}

func TestElasticsearchClient_QueryForESQL_NotConnected(t *testing.T) {
	// 연결되지 않은 클라이언트 (v8)
	client := &ElasticsearchClient{
		version: 8,
		v8:      nil,
	}

	var ctx = context.Background()

	_, err := client.QueryForESQL(ctx, "FROM test", 30)
	assert.Error(t, err)
	assert.Equal(t, external.ErrorNotConnected, err)

	// 연결되지 않은 클라이언트 (v9)
	client9 := &ElasticsearchClient{
		version: 9,
		v9:      nil,
	}

	_, err = client9.QueryForESQL(ctx, "FROM test", 30)
	assert.Error(t, err)
	assert.Equal(t, external.ErrorNotConnected, err)
}

func TestElasticsearchClient_QueryForESQL_UnsupportedVersion(t *testing.T) {
	client := &ElasticsearchClient{
		version: 7,
	}

	var ctx = context.Background()

	_, err := client.QueryForESQL(ctx, "FROM test", 30)
	assert.Error(t, err)
	assert.Equal(t, external.ErrorNotSupported, err)
}

func TestElasticsearchConfig_Validation(t *testing.T) {
	tests := []struct {
		name    string
		config  ElasticsearchConfig
		wantErr bool
	}{
		{
			name: "valid v8 config",
			config: ElasticsearchConfig{
				Version:   8,
				Addresses: []string{"http://localhost:9200"},
				Bulk: struct {
					FlushBytes    int    `mapstructure:"flush_bytes"`
					FlushInterval string `mapstructure:"flush_interval"`
				}{
					FlushBytes:    5000000,
					FlushInterval: "30s",
				},
			},
			wantErr: false,
		},
		{
			name: "valid v9 config",
			config: ElasticsearchConfig{
				Version:   9,
				Addresses: []string{"http://localhost:9200"},
				Bulk: struct {
					FlushBytes    int    `mapstructure:"flush_bytes"`
					FlushInterval string `mapstructure:"flush_interval"`
				}{
					FlushBytes:    5000000,
					FlushInterval: "30s",
				},
			},
			wantErr: false,
		},
		{
			name: "invalid version",
			config: ElasticsearchConfig{
				Version:   6,
				Addresses: []string{"http://localhost:9200"},
				Bulk: struct {
					FlushBytes    int    `mapstructure:"flush_bytes"`
					FlushInterval string `mapstructure:"flush_interval"`
				}{
					FlushBytes:    5000000,
					FlushInterval: "30s",
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, err := NewElasticsearchClient(tt.config)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, client)
			} else {
				if err == nil {
					assert.NotNil(t, client)
					defer client.Close(5)
				}
			}
		})
	}
}

// Benchmark tests
func BenchmarkNewElasticsearchClient(b *testing.B) {
	config := ElasticsearchConfig{
		Version:   8,
		Addresses: []string{"http://localhost:9200"},
		Bulk: struct {
			FlushBytes    int    `mapstructure:"flush_bytes"`
			FlushInterval string `mapstructure:"flush_interval"`
		}{
			FlushBytes:    5000000,
			FlushInterval: "30s",
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		client, err := NewElasticsearchClient(config)
		if err == nil {
			client.Close(5)
		}
	}
}

func BenchmarkElasticsearchClient_GetVersion(b *testing.B) {
	client := &ElasticsearchClient{
		version: 8,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = client.GetVersion()
	}
}
