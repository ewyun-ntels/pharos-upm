package common

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	alertcommon "ntels.com/pharos/core/pkg/alert/common"
	"ntels.com/pharos/core/pkg/common"
	notifcommon "ntels.com/pharos/core/pkg/notification/common"
	notifrule "ntels.com/pharos/core/pkg/notification/rule"
)

type captureRule struct {
	name   string
	mu     sync.Mutex
	values []*notifcommon.AlertValue
	retErr error
}

func (c *captureRule) Load(_ common.Config, _ map[string]any) error { return nil }
func (c *captureRule) GetId() string                                { return c.name }
func (c *captureRule) SetId(_ string)                               {}
func (c *captureRule) GetName() string                              { return c.name }
func (c *captureRule) Validate() error                              { return nil }
func (c *captureRule) Send(v *notifcommon.AlertValue) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.values = append(c.values, v)
	return c.retErr
}
func (c *captureRule) Run(_ context.Context) error { return nil }
func (c *captureRule) Destroy() error              { return nil }

func TestNotificationSender_Send_ForwardsToDestinations(t *testing.T) {
	// Register a capture rule
	c := &captureRule{name: "dest1"}
	if err := notifrule.Register(c); err != nil {
		t.Fatalf("register rule: %v", err)
	}
	defer func() { _ = notifrule.Unregister("dest1") }()

	sender := NotificationSender{Destinations: []string{"dest1"}}

	alerts := []alertcommon.Value{{
		Name:        "rule-A",
		AlertType:   "query",
		Description: "desc",
		AlertId:     "id-1",
		Status:      alertcommon.StatusAlerting,
		Value:       42.5,
		Severity:    SeverityMajor,
		Timestamp:   time.Now(),
		UpdatedAt:   time.Now(),
		Labels:      map[string]string{"k1": "v1"},
	}}

	if err := sender.Send(alerts); err != nil {
		t.Fatalf("Send returned error: %v", err)
	}

	c.mu.Lock()
	defer c.mu.Unlock()
	if len(c.values) != 1 {
		t.Fatalf("expected 1 forwarded value, got %d", len(c.values))
	}
	got := c.values[0]
	if got.Name != "rule-A" || got.Severity != SeverityMajor || got.Labels["k1"] != "v1" {
		t.Fatalf("unexpected forwarded content: %+v", got)
	}
}

func TestNotificationSender_Send_ContinuesOnError(t *testing.T) {
	ok := &captureRule{name: "ok"}
	bad := &captureRule{name: "bad", retErr: errors.New("send failed")}
	if err := notifrule.Register(ok); err != nil {
		t.Fatalf("register ok: %v", err)
	}
	if err := notifrule.Register(bad); err != nil {
		t.Fatalf("register bad: %v", err)
	}
	defer func() { _ = notifrule.Unregister("ok"); _ = notifrule.Unregister("bad") }()

	sender := NotificationSender{Destinations: []string{"bad", "ok"}}

	alerts := []alertcommon.Value{{
		Name:      "n",
		AlertType: "query",
		AlertId:   "x",
		Status:    alertcommon.StatusAlerting,
		Value:     1,
		Severity:  SeverityMinor,
		Timestamp: time.Now(),
		UpdatedAt: time.Now(),
	}}

	// Should not return error even if one destination fails
	if err := sender.Send(alerts); err != nil {
		t.Fatalf("Send should not fail when a destination returns error, got: %v", err)
	}

	// Ensure the ok rule received the value
	ok.mu.Lock()
	defer ok.mu.Unlock()
	if len(ok.values) != 1 {
		t.Fatalf("expected ok rule to receive 1 value, got %d", len(ok.values))
	}
}
