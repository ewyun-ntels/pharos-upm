package repositories

import (
	"testing"

	"ntels.com/pharos/core/external/orm"
)

func Test_BaseRepository_MarshalAttributes(t *testing.T) {
	r := NewBaseRepository(orm.DatabaseConfig{})

	extra, err := r.MarshalAttributes(nil)
	if err != nil {
		t.Errorf("marshalAttributes() error = %v", err)
		return
	}
	expected := []byte("{}")
	if string(extra) != string(expected) {
		t.Errorf("marshalAttributes() = %s, want %s", extra, expected)
	}

	attrs := map[string]any{
		"key1": "value1",
		"key2": 2,
	}
	extra, err = r.MarshalAttributes(attrs)
	if err != nil {
		t.Errorf("marshalAttributes() error = %v", err)
		return
	}
	expectedStr := `{"key1":"value1","key2":2}`
	if string(extra) != expectedStr {
		t.Errorf("marshalAttributes() = %s, want %s", extra, expectedStr)
	}
}

func Test_BaseRepository_UnmarshalAttributes(t *testing.T) {
	r := NewBaseRepository(orm.DatabaseConfig{})

	attrs, err := r.UnmarshalAttributes(nil)
	if err != nil {
		t.Errorf("unmarshalAttributes() error = %v", err)
		return
	}
	if len(attrs) != 0 {
		t.Errorf("unmarshalAttributes() = %v, want empty map", attrs)
	}

	data := []byte(`{"key1":"value1","key2":2}`)
	attrs, err = r.UnmarshalAttributes(data)
	if err != nil {
		t.Errorf("unmarshalAttributes() error = %v", err)
		return
	}
	expected := map[string]any{
		"key1": "value1",
		"key2": float64(2), // JSON 숫자는 기본적으로 float64로 언마샬링됨
	}
	if len(attrs) != len(expected) {
		t.Errorf("unmarshalAttributes() = %v, want %v", attrs, expected)
		return
	}
	for k, v := range expected {
		if attrs[k] != v {
			t.Errorf("unmarshalAttributes()[%s] = %v, want %v", k, attrs[k], v)
		}
	}
}

func Test_BaseRepository_MarshalPrepare(t *testing.T) {
	r := NewBaseRepository(orm.DatabaseConfig{})

	prepare, err := r.MarshalPrepare(nil)
	if err != nil {
		t.Errorf("marshalPrepare() error = %v", err)
		return
	}
	expected := []byte("[]")
	if string(prepare) != string(expected) {
		t.Errorf("marshalPrepare() = %s, want %s", prepare, expected)
	}

	prepList := []string{"cmd1", "cmd2"}
	prepare, err = r.MarshalPrepare(prepList)
	if err != nil {
		t.Errorf("marshalPrepare() error = %v", err)
		return
	}
	expectedStr := `["cmd1","cmd2"]`
	if string(prepare) != expectedStr {
		t.Errorf("marshalPrepare() = %s, want %s", prepare, expectedStr)
	}
}

func Test_BaseRepository_UnmarshalPrepare(t *testing.T) {
	r := NewBaseRepository(orm.DatabaseConfig{})

	prepare, err := r.UnmarshalPrepare(nil)
	if err != nil {
		t.Errorf("unmarshalPrepare() error = %v", err)
		return
	}
	if len(prepare) != 0 {
		t.Errorf("unmarshalPrepare() = %v, want empty slice", prepare)
	}

	data := []byte(`["cmd1","cmd2"]`)
	prepare, err = r.UnmarshalPrepare(data)
	if err != nil {
		t.Errorf("unmarshalPrepare() error = %v", err)
		return
	}
	expected := []string{"cmd1", "cmd2"}
	if len(prepare) != len(expected) {
		t.Errorf("unmarshalPrepare() = %v, want %v", prepare, expected)
		return
	}
	for i, v := range expected {
		if prepare[i] != v {
			t.Errorf("unmarshalPrepare()[%d] = %v, want %v", i, prepare[i], v)
		}
	}
}
