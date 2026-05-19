package metrics

import (
	"testing"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/assert"
)

// Mock collector for testing
type mockCollector struct {
	describeCalled bool
	collectCalled  bool
}

func (m *mockCollector) Describe(ch chan<- *prometheus.Desc) {
	m.describeCalled = true
}

func (m *mockCollector) Collect(ch chan<- prometheus.Metric) {
	m.collectCalled = true
}

func TestAddCollector(t *testing.T) {
	mock := &mockCollector{}
	AddCollector("test_collector", mock)

	// Verify collector was added
	collector := collectors.Get("test_collector")
	assert.NotNil(t, collector)
	assert.Equal(t, mock, collector)
}

func TestAddCollector_Multiple(t *testing.T) {
	mock1 := &mockCollector{}
	mock2 := &mockCollector{}
	mock3 := &mockCollector{}

	AddCollector("collector1", mock1)
	AddCollector("collector2", mock2)
	AddCollector("collector3", mock3)

	// Verify all collectors were added
	assert.Equal(t, mock1, collectors.Get("collector1"))
	assert.Equal(t, mock2, collectors.Get("collector2"))
	assert.Equal(t, mock3, collectors.Get("collector3"))

	// Verify GetAll returns all collectors
	allCollectors := collectors.GetAll()
	assert.GreaterOrEqual(t, len(allCollectors), 3)
	assert.Contains(t, allCollectors, "collector1")
	assert.Contains(t, allCollectors, "collector2")
	assert.Contains(t, allCollectors, "collector3")
}

func TestAddCollector_Overwrite(t *testing.T) {
	mock1 := &mockCollector{}
	mock2 := &mockCollector{}

	AddCollector("test", mock1)
	assert.Equal(t, mock1, collectors.Get("test"))

	// Overwrite with new collector
	AddCollector("test", mock2)
	assert.Equal(t, mock2, collectors.Get("test"))
}

func TestCollectors_GetAll(t *testing.T) {
	// Add multiple collectors
	AddCollector("col1", &mockCollector{})
	AddCollector("col2", &mockCollector{})

	allCollectors := collectors.GetAll()
	assert.NotNil(t, allCollectors)
	assert.IsType(t, map[string]prometheus.Collector{}, allCollectors)
}

func TestCollectors_ThreadSafety(t *testing.T) {
	// Test concurrent access
	done := make(chan bool)

	// Writer goroutine
	go func() {
		for range 100 {
			AddCollector("concurrent", &mockCollector{})
		}
		done <- true
	}()

	// Reader goroutine
	go func() {
		for range 100 {
			_ = collectors.Get("concurrent")
		}
		done <- true
	}()

	// Wait for both
	<-done
	<-done

	// Verify no panic occurred
	assert.NotNil(t, collectors)
}

func BenchmarkAddCollector(b *testing.B) {
	mock := &mockCollector{}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		AddCollector("bench", mock)
	}
}

func BenchmarkGetCollector(b *testing.B) {
	mock := &mockCollector{}
	AddCollector("bench", mock)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = collectors.Get("bench")
	}
}

func BenchmarkGetAllCollectors(b *testing.B) {
	// Add some collectors
	for range 10 {
		AddCollector("bench", &mockCollector{})
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = collectors.GetAll()
	}
}
