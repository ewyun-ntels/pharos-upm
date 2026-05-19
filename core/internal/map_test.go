package internal

import (
	"reflect"
	"slices"
	"sync"
	"testing"
)

func TestNewMap(t *testing.T) {
	// Test creating a new map with string values
	stringMap := NewMap[string]()
	if stringMap == nil {
		t.Fatal("NewMap should not return nil")
	}
	if stringMap.datas == nil {
		t.Fatal("NewMap should initialize datas map")
	}
	if len(stringMap.datas) != 0 {
		t.Fatal("NewMap should create empty map")
	}

	// Test creating a new map with int values
	intMap := NewMap[int]()
	if intMap == nil {
		t.Fatal("NewMap should not return nil")
	}
}

func TestMap_Set_Get(t *testing.T) {
	m := NewMap[string]()

	// Test setting and getting a value
	key := "test_key"
	value := "test_value"
	m.Set(key, value)

	result := m.Get(key)
	if result != value {
		t.Errorf("Expected %s, got %s", value, result)
	}

	// Test getting non-existent key (should return zero value)
	nonExistentKey := "non_existent"
	result = m.Get(nonExistentKey)
	if result != "" {
		t.Errorf("Expected empty string for non-existent key, got %s", result)
	}
}

func TestMap_Exist(t *testing.T) {
	m := NewMap[int]()

	// Test non-existent key
	if m.Exist("key1") {
		t.Error("Exist should return false for non-existent key")
	}

	// Test existing key
	m.Set("key1", 42)
	if !m.Exist("key1") {
		t.Error("Exist should return true for existing key")
	}

	// Test after removal
	m.Remove("key1", nil)
	if m.Exist("key1") {
		t.Error("Exist should return false after removal")
	}
}

func TestMap_GetAll(t *testing.T) {
	m := NewMap[string]()

	// Test empty map
	all := m.GetAll()
	if len(all) != 0 {
		t.Error("GetAll should return empty map for new map")
	}

	// Test with data
	m.Set("key1", "value1")
	m.Set("key2", "value2")
	m.Set("key3", "value3")

	all = m.GetAll()
	expected := map[string]string{
		"key1": "value1",
		"key2": "value2",
		"key3": "value3",
	}

	if !reflect.DeepEqual(all, expected) {
		t.Errorf("GetAll returned incorrect map. Expected %v, got %v", expected, all)
	}
}

func TestMap_Remove(t *testing.T) {
	m := NewMap[string]()

	// Test removing non-existent key
	m.Remove("non_existent", nil)
	// Should not panic or cause issues

	// Test removing existing key without finalFunc
	m.Set("key1", "value1")
	m.Remove("key1", nil)
	if m.Exist("key1") {
		t.Error("Key should be removed")
	}

	// Test removing existing key with finalFunc
	m.Set("key2", "value2")
	var finalFuncCalled bool
	var finalFuncValue string

	finalFunc := func(t string) {
		finalFuncCalled = true
		finalFuncValue = t
	}

	m.Remove("key2", finalFunc)

	if !finalFuncCalled {
		t.Error("Final function should be called")
	}
	if finalFuncValue != "value2" {
		t.Errorf("Final function should receive correct value. Expected 'value2', got '%s'", finalFuncValue)
	}
	if m.Exist("key2") {
		t.Error("Key should be removed")
	}
}

func TestMap_RemoveAll(t *testing.T) {
	m := NewMap[int]()

	// Test removing all from empty map
	m.RemoveAll(nil)
	// Should not panic

	// Test removing all without finalFunc
	m.Set("key1", 1)
	m.Set("key2", 2)
	m.Set("key3", 3)

	m.RemoveAll(nil)

	if len(m.GetAll()) != 0 {
		t.Error("All keys should be removed")
	}

	// Test removing all with finalFunc
	m.Set("key1", 10)
	m.Set("key2", 20)
	m.Set("key3", 30)

	var finalFuncCallCount int
	var finalFuncValues []int

	finalFunc := func(t int) {
		finalFuncCallCount++
		finalFuncValues = append(finalFuncValues, t)
	}

	m.RemoveAll(finalFunc)

	if finalFuncCallCount != 3 {
		t.Errorf("Final function should be called 3 times, called %d times", finalFuncCallCount)
	}

	expectedValues := []int{10, 20, 30}
	if len(finalFuncValues) != len(expectedValues) {
		t.Errorf("Final function should receive all values. Expected %d values, got %d", len(expectedValues), len(finalFuncValues))
	}

	// Check if all values were passed to finalFunc (order may vary due to map iteration)
	for _, expected := range expectedValues {
		found := slices.Contains(finalFuncValues, expected)
		if !found {
			t.Errorf("Value %d should be passed to final function", expected)
		}
	}

	if len(m.GetAll()) != 0 {
		t.Error("All keys should be removed")
	}
}

func TestMap_ConcurrentAccess(t *testing.T) {
	m := NewMap[int]()

	// Test concurrent access to ensure thread safety
	const numGoroutines = 100
	const numOperations = 1000

	var wg sync.WaitGroup
	wg.Add(numGoroutines)

	// Start multiple goroutines performing various operations
	for i := range numGoroutines {
		go func(goroutineID int) {
			defer wg.Done()

			for j := range numOperations {
				key := "key_" + string(rune(goroutineID))
				value := goroutineID*numOperations + j

				// Perform various operations
				m.Set(key, value)
				m.Get(key)
				m.Exist(key)
				m.GetAll()

				if j%10 == 0 {
					m.Remove(key, nil)
				}
			}
		}(i)
	}

	wg.Wait()

	// Test should complete without race conditions or deadlocks
	// The exact final state is not predictable due to concurrent access,
	// but the test verifies that no panics or race conditions occur
}

func TestMap_TypeSafety(t *testing.T) {
	// Test with different types to ensure type safety

	// String map
	stringMap := NewMap[string]()
	stringMap.Set("key1", "string_value")
	if stringMap.Get("key1") != "string_value" {
		t.Error("String map should handle string values correctly")
	}

	// Int map
	intMap := NewMap[int]()
	intMap.Set("key1", 42)
	if intMap.Get("key1") != 42 {
		t.Error("Int map should handle int values correctly")
	}

	// Struct map
	type TestStruct struct {
		Name string
		Age  int
	}

	structMap := NewMap[TestStruct]()
	testStruct := TestStruct{Name: "Test", Age: 25}
	structMap.Set("key1", testStruct)

	result := structMap.Get("key1")
	if result.Name != "Test" || result.Age != 25 {
		t.Error("Struct map should handle struct values correctly")
	}

	// Pointer map
	pointerMap := NewMap[*TestStruct]()
	testStructPtr := &TestStruct{Name: "TestPtr", Age: 30}
	pointerMap.Set("key1", testStructPtr)

	resultPtr := pointerMap.Get("key1")
	if resultPtr == nil || resultPtr.Name != "TestPtr" || resultPtr.Age != 30 {
		t.Error("Pointer map should handle pointer values correctly")
	}
}

func TestMap_ZeroValues(t *testing.T) {
	// Test behavior with zero values

	// String map with empty string
	stringMap := NewMap[string]()
	stringMap.Set("empty", "")
	if !stringMap.Exist("empty") {
		t.Error("Map should store empty string")
	}
	if stringMap.Get("empty") != "" {
		t.Error("Map should return empty string")
	}

	// Int map with zero
	intMap := NewMap[int]()
	intMap.Set("zero", 0)
	if !intMap.Exist("zero") {
		t.Error("Map should store zero value")
	}
	if intMap.Get("zero") != 0 {
		t.Error("Map should return zero value")
	}

	// Pointer map with nil
	pointerMap := NewMap[*string]()
	pointerMap.Set("nil", nil)
	if !pointerMap.Exist("nil") {
		t.Error("Map should store nil pointer")
	}
	if pointerMap.Get("nil") != nil {
		t.Error("Map should return nil pointer")
	}
}

// Benchmark tests
func BenchmarkMap_Set(b *testing.B) {
	m := NewMap[int]()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		m.Set("key", i)
	}
}

func BenchmarkMap_Get(b *testing.B) {
	m := NewMap[int]()
	m.Set("key", 42)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		m.Get("key")
	}
}

func BenchmarkMap_Exist(b *testing.B) {
	m := NewMap[int]()
	m.Set("key", 42)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		m.Exist("key")
	}
}

func BenchmarkMap_ConcurrentAccess(b *testing.B) {
	m := NewMap[int]()

	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			key := "key_" + string(rune(i%100))
			m.Set(key, i)
			m.Get(key)
			i++
		}
	})
}
