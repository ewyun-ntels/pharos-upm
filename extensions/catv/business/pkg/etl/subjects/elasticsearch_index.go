package subjects

import (
	"log/slog"
	"regexp"
	"strings"

	"github.com/nats-io/nats.go"
	"ntels.com/pharos/core/external/orm"
	"ntels.com/pharos/core/pkg/common"
	"ntels.com/pharos/core/pkg/message_router"
)

// isValidIndexName은 Elasticsearch 인덱스 이름이 유효한지 검증합니다.
func isValidIndexName(index string) bool {
	if index == "" || len(index) > 255 {
		return false
	}
	// Elasticsearch 인덱스 이름 규칙: 소문자, 숫자, -, _, + 만 허용
	// -, _, + 로 시작할 수 없음
	matched, _ := regexp.MatchString(`^[a-z0-9][a-z0-9_\-+]*$`, index)
	return matched
}

func GetElasticsearchIndexSubject(config common.Config) message_router.Subject {
	const subject = "elasticsearch.index"

	return message_router.Subject{
		Key: message_router.ElasticsearchIndex,

		RequestSubject: subject,
		RequestHandler: func(message message_router.Message) {
			slog.Debug("receive message", "Subject", message.Subject, "Reply", message.Reply, "Data", string(message.Data))

			msg, ok := message.Raw.(*nats.Msg)
			if !ok {
				slog.Error("invalid raw message type")
				return
			}

			datas := strings.Split(string(msg.Data), "\t")
			if len(datas) < 3 {
				slog.Error("invalid data format", "data", string(msg.Data))
				return
			}

			index := datas[0]
			documentID := datas[1]
			document := datas[2]

			if !isValidIndexName(index) {
				slog.Error("invalid index name", "index", index)
				return
			}

			client, err := orm.DatabasePool.GetElasticsearchClient(orm.DatabaseConfig{Driver: orm.DriverElasticsearch, Elasticsearch: config.ETL.Elasticsearch})
			if err != nil {
				slog.Error("failed to get Elasticsearch client", "error", err)
				return
			}

			if err := client.IndexBulk(index, documentID, document); err != nil {
				slog.Error("failed to index document", "error", err)
				return
			}
		},

		ReplySubject: "",
		ReplyHandler: nil,
	}
}
