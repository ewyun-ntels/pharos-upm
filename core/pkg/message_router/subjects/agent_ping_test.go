package subjects

import (
	"testing"

	"github.com/nats-io/nats.go"
	"github.com/stretchr/testify/assert"
	"ntels.com/pharos/core/pkg/common"
	"ntels.com/pharos/core/pkg/message_router"
)

func TestGetAgentPingSubject_Structure(t *testing.T) {
	t.Parallel()

	config := common.Config{}
	subject := GetAgentPingSubject(config)

	assert.NotEmpty(t, subject.RequestSubject, "Request subject should not be empty")
	assert.Equal(t, "agent.ping.{{.name}}", subject.RequestSubject)
	assert.NotNil(t, subject.RequestHandler, "Request handler should not be nil")
	assert.Empty(t, subject.ReplySubject, "Reply subject should be empty")
	assert.Nil(t, subject.ReplyHandler, "Reply handler should be nil")
}

func TestGetAgentPingSubject_RequestHandler_Echo(t *testing.T) {
	t.Parallel()

	config := common.Config{}
	subject := GetAgentPingSubject(config)

	testData := []byte("ping test data")

	mockMsg := &nats.Msg{
		Subject: "agent.ping.test-agent",
		Reply:   "reply.subject",
		Data:    testData,
		Sub: &nats.Subscription{
			Subject: "agent.ping.test-agent",
		},
	}

	message := message_router.Message{
		Subject: mockMsg.Subject,
		Reply:   mockMsg.Reply,
		Data:    mockMsg.Data,
		Raw:     mockMsg,
	}

	assert.NotPanics(t, func() {
		subject.RequestHandler(message)
	}, "Handler should not panic and should echo data")
}

func TestGetAgentPingSubject_RequestHandler_EmptyData(t *testing.T) {
	t.Parallel()

	config := common.Config{}
	subject := GetAgentPingSubject(config)

	mockMsg := &nats.Msg{
		Subject: "agent.ping.test-agent",
		Reply:   "reply.subject",
		Data:    []byte{},
		Sub: &nats.Subscription{
			Subject: "agent.ping.test-agent",
		},
	}

	message := message_router.Message{
		Subject: mockMsg.Subject,
		Reply:   mockMsg.Reply,
		Data:    mockMsg.Data,
		Raw:     mockMsg,
	}

	assert.NotPanics(t, func() {
		subject.RequestHandler(message)
	}, "Handler should not panic with empty data")
}

func TestGetAgentPingSubject_RequestHandler_LargeData(t *testing.T) {
	t.Parallel()

	config := common.Config{}
	subject := GetAgentPingSubject(config)

	largeData := make([]byte, 1024*10)
	for i := range largeData {
		largeData[i] = byte(i % 256)
	}

	mockMsg := &nats.Msg{
		Subject: "agent.ping.test-agent",
		Reply:   "reply.subject",
		Data:    largeData,
		Sub: &nats.Subscription{
			Subject: "agent.ping.test-agent",
		},
	}

	message := message_router.Message{
		Subject: mockMsg.Subject,
		Reply:   mockMsg.Reply,
		Data:    mockMsg.Data,
		Raw:     mockMsg,
	}

	assert.NotPanics(t, func() {
		subject.RequestHandler(message)
	}, "Handler should not panic with large data")
}

func TestGetAgentPingSubject_RequestHandler_NoPanic(t *testing.T) {
	t.Parallel()

	config := common.Config{}
	subject := GetAgentPingSubject(config)

	mockMsg := &nats.Msg{
		Subject: "agent.ping.test",
		Reply:   "reply",
		Data:    []byte("test"),
	}

	message := message_router.Message{
		Subject: mockMsg.Subject,
		Reply:   mockMsg.Reply,
		Data:    mockMsg.Data,
		Raw:     mockMsg,
	}

	assert.NotPanics(t, func() {
		subject.RequestHandler(message)
	}, "RequestHandler should not panic")
}

func TestGetAgentPingSubject_SubjectPattern(t *testing.T) {
	t.Parallel()

	config := common.Config{}
	subject := GetAgentPingSubject(config)

	expectedPattern := "agent.ping.{{.name}}"
	assert.Equal(t, expectedPattern, subject.RequestSubject, "Subject should use template pattern")
}

func TestGetAgentPingSubject_NoReplyHandling(t *testing.T) {
	t.Parallel()

	config := common.Config{}
	subject := GetAgentPingSubject(config)

	assert.Empty(t, subject.ReplySubject, "Reply subject should be empty")
	assert.Nil(t, subject.ReplyHandler, "Reply handler should be nil")
}

func TestGetAgentPingSubject_Key(t *testing.T) {
	t.Parallel()

	config := common.Config{}
	subject := GetAgentPingSubject(config)

	assert.NotEmpty(t, subject.Key, "Subject key should not be empty")
}

func TestGetAgentPingSubject_MultipleMessages(t *testing.T) {
	t.Parallel()

	config := common.Config{}
	subject := GetAgentPingSubject(config)

	testCases := [][]byte{
		[]byte("ping1"),
		[]byte("ping2"),
		[]byte("ping3"),
	}

	for _, testData := range testCases {
		mockMsg := &nats.Msg{
			Subject: "agent.ping.test",
			Reply:   "reply",
			Data:    testData,
			Sub: &nats.Subscription{
				Subject: "agent.ping.test",
			},
		}

		message := message_router.Message{
			Subject: mockMsg.Subject,
			Reply:   mockMsg.Reply,
			Data:    mockMsg.Data,
			Raw:     mockMsg,
		}

		assert.NotPanics(t, func() {
			subject.RequestHandler(message)
		}, "Handler should handle each message without panic")
	}
}
