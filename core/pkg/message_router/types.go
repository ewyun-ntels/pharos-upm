package message_router

import (
	"os"

	"github.com/google/uuid"
	"ntels.com/pharos/core/internal"
)

var (
	AgentPing = uuid.New().String()

	ElasticsearchIndex = uuid.New().String()

	ReplySuffixSpecific = ".reply.{{.name}}"
	ReplySuffixAll      = ".reply"
)

// Message는 클라이언트 독립적인 공통 메시지 구조체입니다.
type Message struct {
	// 메시지 데이터
	Data []byte

	// 메시지 주제/토픽
	Subject string

	// 응답 주제 (Request-Reply 패턴용)
	Reply string

	// 메시지 헤더 (key-value)
	Headers map[string][]string

	// 원본 메시지 (필요시 타입 단언으로 접근)
	// NATS: *nats.Msg, Kafka: *kafka.Message 등
	Raw any
}

// MessageHandler는 클라이언트 독립적인 공통 메시지 핸들러입니다.
// 구독한 토픽/주제에서 메시지를 수신할 때 호출되는 콜백 함수입니다.
type MessageHandler func(msg Message)

type Subject struct {
	Key string

	RequestSubject string
	RequestHandler MessageHandler

	ReplySubject string
	ReplyHandler MessageHandler
}

var Subjects = internal.NewMap[Subject]()
var AddSubjects = internal.NewMap[Subject]()

func AddSubject(subject Subject) {
	AddSubjects.Set(subject.Key, subject)
}

func GetName() string {
	hostname, _ := os.Hostname()
	return hostname
}
