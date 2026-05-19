package rule

import (
	"context"
	"errors"
	"testing"

	commonpkg "ntels.com/pharos/core/pkg/common"
	notifcommon "ntels.com/pharos/core/pkg/notification/common"
)

type fakeRule struct {
	id       string
	name     string
	sendErr  error
	sentWith *notifcommon.AlertValue
}

func (f *fakeRule) Load(_ commonpkg.Config, _ map[string]any) error { return nil }
func (f *fakeRule) GetId() string                                   { return f.id }
func (f *fakeRule) SetId(id string)                                 { f.id = id }
func (f *fakeRule) GetName() string                                 { return f.name }
func (f *fakeRule) Validate() error                                 { return nil }
func (f *fakeRule) Send(v *notifcommon.AlertValue) error {
	f.sentWith = v
	return f.sendErr
}
func (f *fakeRule) Run(_ context.Context) error { return nil }
func (f *fakeRule) Destroy() error              { return nil }

func resetManager() {
	gManager = Manager{Rules: make(map[string]Rule)}
}

func Test_Manager_RegisterAndDuplicate(t *testing.T) {
	resetManager()
	fr := &fakeRule{name: "r1"}
	if err := Register(fr); err != nil {
		t.Fatalf("unexpected error registering: %v", err)
	}
	// duplicate name
	if err := Register(&fakeRule{name: "r1"}); err == nil {
		t.Fatalf("expected duplicate registration error")
	}
}

func Test_Manager_Send_NotFound(t *testing.T) {
	resetManager()
	if err := Send("nope", &notifcommon.AlertValue{Name: "n"}); err == nil {
		t.Fatalf("expected error for missing rule")
	}
}

func Test_Manager_Send_PropagatesToRule(t *testing.T) {
	resetManager()
	fr := &fakeRule{name: "r1"}
	if err := Register(fr); err != nil {
		t.Fatalf("register err: %v", err)
	}
	val := &notifcommon.AlertValue{Name: "n"}
	if err := Send("r1", val); err != nil {
		t.Fatalf("unexpected send err: %v", err)
	}
	if fr.sentWith != val {
		t.Fatalf("rule did not receive value")
	}
}

func Test_Manager_Send_PropagateError(t *testing.T) {
	resetManager()
	expected := errors.New("boom")
	fr := &fakeRule{name: "r1", sendErr: expected}
	if err := Register(fr); err != nil {
		t.Fatalf("register err: %v", err)
	}
	err := Send("r1", &notifcommon.AlertValue{Name: "n"})
	if err == nil || err.Error() != expected.Error() {
		t.Fatalf("expected error propagation, got %v", err)
	}
}

func Test_Manager_Unregister(t *testing.T) {
	resetManager()
	_ = Register(&fakeRule{name: "r1"})
	if _, ok := gManager.Rules["r1"]; !ok {
		t.Fatalf("rule not registered")
	}
	_ = Unregister("r1")
	if _, ok := gManager.Rules["r1"]; ok {
		t.Fatalf("rule still present after unregister")
	}
}
