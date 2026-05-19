package subjects

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"ntels.com/pharos/core/pkg/common"
	"ntels.com/pharos/core/pkg/message_router"
)

// TestIsValidIndexName은 Elasticsearch 인덱스 이름 검증 로직을 테스트합니다.
// Elasticsearch의 인덱스 명명 규칙:
// - 소문자, 숫자, -, _, + 만 허용
// - -, _, + 로 시작할 수 없음
// - 최대 255자
func TestIsValidIndexName(t *testing.T) {
	tests := []struct {
		name      string
		indexName string
		want      bool
	}{
		{
			name:      "valid lowercase",
			indexName: "myindex",
			want:      true,
		},
		{
			name:      "valid with numbers",
			indexName: "index123",
			want:      true,
		},
		{
			name:      "valid with hyphen",
			indexName: "my-index",
			want:      true,
		},
		{
			name:      "valid with underscore",
			indexName: "my_index",
			want:      true,
		},
		{
			name:      "valid with plus",
			indexName: "my+index",
			want:      true,
		},
		{
			name:      "invalid empty",
			indexName: "",
			want:      false,
		},
		{
			name:      "invalid uppercase",
			indexName: "MyIndex",
			want:      false,
		},
		{
			name:      "invalid starts with hyphen",
			indexName: "-myindex",
			want:      false,
		},
		{
			name:      "invalid starts with underscore",
			indexName: "_myindex",
			want:      false,
		},
		{
			name:      "invalid starts with plus",
			indexName: "+myindex",
			want:      false,
		},
		{
			name:      "invalid special characters",
			indexName: "my@index",
			want:      false,
		},
		{
			name:      "invalid with space",
			indexName: "my index",
			want:      false,
		},
		{
			name:      "invalid too long",
			indexName: string(make([]byte, 256)),
			want:      false,
		},
		{
			name:      "valid max length",
			indexName: string(make([]byte, 255)),
			want:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 255자와 256자 테스트의 경우 유효한 문자로 채움
			indexName := tt.indexName
			switch tt.name {
			case "invalid too long":
				// 256자의 유효한 문자열 생성
				indexName = "a" + strings.Repeat("x", 255)
			case "valid max length":
				// 255자의 유효한 문자열 생성
				indexName = "a" + strings.Repeat("x", 254)
			}

			got := isValidIndexName(indexName)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestGetElasticsearchIndexSubject(t *testing.T) {
	config := common.Config{
		Serve: common.ServeConfig{},
	}

	subject := GetElasticsearchIndexSubject(config)

	// Subject 기본 속성 검증
	assert.Equal(t, message_router.ElasticsearchIndex, subject.Key)
	assert.Equal(t, "elasticsearch.index", subject.RequestSubject)
	assert.NotNil(t, subject.RequestHandler)
	assert.Empty(t, subject.ReplySubject)
	assert.Nil(t, subject.ReplyHandler)
}

func TestGetElasticsearchIndexSubject_Structure(t *testing.T) {
	config := common.Config{}
	subject := GetElasticsearchIndexSubject(config)

	// Subject가 올바른 구조를 가지고 있는지 확인
	require.NotEmpty(t, subject.Key)
	require.NotEmpty(t, subject.RequestSubject)
	require.NotNil(t, subject.RequestHandler)
}

func TestIsValidIndexName_EdgeCases(t *testing.T) {
	// 경계 조건 테스트
	tests := []struct {
		name      string
		indexName string
		want      bool
	}{
		{
			name:      "single character valid",
			indexName: "a",
			want:      true,
		},
		{
			name:      "single number valid",
			indexName: "0",
			want:      true,
		},
		{
			name:      "numbers only",
			indexName: "123456",
			want:      true,
		},
		{
			name:      "complex valid name",
			indexName: "log-2024-01-01_data+backup",
			want:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isValidIndexName(tt.indexName)
			assert.Equal(t, tt.want, got)
		})
	}
}

// Benchmark tests
func BenchmarkIsValidIndexName(b *testing.B) {
	indexNames := []string{
		"myindex",
		"my-index-123",
		"invalid@index",
		"_invalid",
	}

	for i := 0; i < b.N; i++ {
		for _, name := range indexNames {
			isValidIndexName(name)
		}
	}
}

func BenchmarkGetElasticsearchIndexSubject(b *testing.B) {
	config := common.Config{
		Serve: common.ServeConfig{},
	}

	for i := 0; i < b.N; i++ {
		GetElasticsearchIndexSubject(config)
	}
}
