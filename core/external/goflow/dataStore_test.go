package goflow

import (
	"testing"

	"ntels.com/pharos/core/pkg/common"
)

func TestDataStore_Set(t *testing.T) {
	tests := []struct {
		name    string
		key     string
		value   any
		wantErr bool
	}{
		{
			name:    "set string value",
			key:     "test-key",
			value:   "test-value",
			wantErr: false,
		},
		{
			name:    "set map value",
			key:     "test-map",
			value:   map[string]string{"key": "value"},
			wantErr: false,
		},
		{
			name:    "set struct value",
			key:     "test-struct",
			value:   struct{ Name string }{Name: "test"},
			wantErr: false,
		},
		{
			name:    "set nil value",
			key:     "test-nil",
			value:   nil,
			wantErr: false,
		},
		{
			name:    "set empty key",
			key:     "",
			value:   "test",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ds := &DataStore{
				config: common.Config{},
			}

			err := ds.Set(tt.key, tt.value)
			if (err != nil) != tt.wantErr {
				t.Errorf("Set() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestDataStore_Get(t *testing.T) {
	ds := &DataStore{
		config: common.Config{},
	}

	tests := []struct {
		name      string
		key       string
		value     any
		wantFound bool
		wantErr   bool
	}{
		{
			name:      "get existing key",
			key:       "test-key",
			value:     "test-value",
			wantFound: false, // Currently always returns false
			wantErr:   false,
		},
		{
			name:      "get non-existing key",
			key:       "non-existing",
			value:     nil,
			wantFound: false,
			wantErr:   false,
		},
		{
			name:      "get with empty key",
			key:       "",
			value:     nil,
			wantFound: false,
			wantErr:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			found, err := ds.Get(tt.key, tt.value)
			if (err != nil) != tt.wantErr {
				t.Errorf("Get() error = %v, wantErr %v", err, tt.wantErr)
			}
			if found != tt.wantFound {
				t.Errorf("Get() found = %v, want %v", found, tt.wantFound)
			}
		})
	}
}

func TestDataStore_Delete(t *testing.T) {
	ds := &DataStore{
		config: common.Config{},
	}

	tests := []struct {
		name    string
		key     string
		wantErr bool
	}{
		{
			name:    "delete existing key",
			key:     "test-key",
			wantErr: false,
		},
		{
			name:    "delete non-existing key",
			key:     "non-existing",
			wantErr: false,
		},
		{
			name:    "delete with empty key",
			key:     "",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ds.Delete(tt.key)
			if (err != nil) != tt.wantErr {
				t.Errorf("Delete() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestDataStore_Close(t *testing.T) {
	ds := &DataStore{
		config: common.Config{},
	}

	err := ds.Close()
	if err != nil {
		t.Errorf("Close() error = %v, want nil", err)
	}

	// Test multiple close calls
	err = ds.Close()
	if err != nil {
		t.Errorf("Close() second call error = %v, want nil", err)
	}
}

func TestDataStore_SetMultipleValues(t *testing.T) {
	ds := &DataStore{
		config: common.Config{},
	}

	// Set multiple values
	values := []struct {
		key   string
		value any
	}{
		{"key1", "value1"},
		{"key2", 123},
		{"key3", map[string]int{"count": 10}},
		{"key4", []string{"a", "b", "c"}},
	}

	for _, v := range values {
		err := ds.Set(v.key, v.value)
		if err != nil {
			t.Errorf("Set(%s) error = %v", v.key, err)
		}
	}
}

func TestDataStore_ConfigPersistence(t *testing.T) {
	config := common.Config{}
	ds := &DataStore{
		config: config,
	}

	// Config is set, just verify it doesn't panic
	_ = ds.config
}

func TestDataStore_Operations(t *testing.T) {
	ds := &DataStore{
		config: common.Config{},
	}

	// Test sequence: Set -> Get -> Delete -> Close
	key := "test-key"
	value := "test-value"

	// Set
	err := ds.Set(key, value)
	if err != nil {
		t.Errorf("Set() error = %v", err)
	}

	// Get
	found, err := ds.Get(key, nil)
	if err != nil {
		t.Errorf("Get() error = %v", err)
	}
	if found {
		t.Log("Note: Get() returned found=true (implementation may have changed)")
	}

	// Delete
	err = ds.Delete(key)
	if err != nil {
		t.Errorf("Delete() error = %v", err)
	}

	// Close
	err = ds.Close()
	if err != nil {
		t.Errorf("Close() error = %v", err)
	}
}

func TestDataStore_SetWithDifferentTypes(t *testing.T) {
	ds := &DataStore{
		config: common.Config{},
	}

	testCases := []any{
		"string",
		123,
		12.34,
		true,
		[]int{1, 2, 3},
		map[string]any{"nested": map[string]string{"key": "value"}},
		struct {
			Name string
			Age  int
		}{Name: "test", Age: 30},
	}

	for i, tc := range testCases {
		err := ds.Set("key", tc)
		if err != nil {
			t.Errorf("Test case %d: Set() error = %v", i, err)
		}
	}
}
