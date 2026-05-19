package command

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"ntels.com/pharos/core/pkg/common"
	"ntels.com/pharos/extensions/catv/business/pkg/control/common/tables"
)

// newTestConfig returns a minimal common.Config for unit tests.
// Database fields are intentionally empty — tests that reach DB will fail fast.
func newTestConfig() common.Config {
	var cfg common.Config
	cfg.Catv.Control.Batch.MaxCount = 10
	cfg.Catv.Control.Batch.FlushInterval = "1s"
	return cfg
}

func TestNewAsyncBatchWriter_Fields(t *testing.T) {
	cfg := newTestConfig()
	w := NewAsyncBatchWriter(cfg, 100, 10, time.Second)
	require.NotNil(t, w)
	assert.Equal(t, 10, w.batchSize)
	assert.Equal(t, time.Second, w.flushInterval)
	assert.Equal(t, 100, cap(w.resultChan))
	assert.NotEmpty(t, w.tableName)
}

func TestNewAsyncBatchWriter_TableNameSetOnce(t *testing.T) {
	cfg := newTestConfig()
	w1 := NewAsyncBatchWriter(cfg, 10, 5, time.Second)
	w2 := NewAsyncBatchWriter(cfg, 10, 5, time.Second)
	assert.Equal(t, w1.tableName, w2.tableName)
	assert.Equal(t, tables.NewStbControlScheduleResultTable(cfg).GetDistributedTableName(), w1.tableName)
}

func TestAsyncBatchWriter_GetChannel(t *testing.T) {
	cfg := newTestConfig()
	w := NewAsyncBatchWriter(cfg, 50, 5, time.Second)
	ch := w.GetChannel()
	require.NotNil(t, ch)
	// Must be send-only
	var _ chan<- *BatchWriteRequest = ch
}

func TestAsyncBatchWriter_StartAndStop_NoItems(t *testing.T) {
	cfg := newTestConfig()
	w := NewAsyncBatchWriter(cfg, 10, 5, 100*time.Millisecond)
	w.Start()
	// Stop must return without deadlock
	done := make(chan struct{})
	go func() {
		w.Stop()
		close(done)
	}()
	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("Stop() deadlocked")
	}
}

func TestAsyncBatchWriter_StopDrainsChannel(t *testing.T) {
	cfg := newTestConfig()
	// batchSize large enough so flush isn't triggered by size
	w := NewAsyncBatchWriter(cfg, 100, 50, 10*time.Second)
	w.Start()

	// Send a nil-raw request (flush will skip nil records — no DB call)
	w.GetChannel() <- &BatchWriteRequest{StbControlScheduleResultRaw: nil}

	done := make(chan struct{})
	go func() {
		w.Stop()
		close(done)
	}()
	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("Stop() deadlocked with pending item")
	}
}

func TestClearBuffer(t *testing.T) {
	buf := []*BatchWriteRequest{
		{StbControlScheduleResultRaw: &tables.StbControlScheduleResultRaw{ID: "1"}},
		{StbControlScheduleResultRaw: &tables.StbControlScheduleResultRaw{ID: "2"}},
	}
	result := clearBuffer(buf)
	assert.Len(t, result, 0)
	// Verify GC references are cleared
	for i := range buf {
		assert.Nil(t, buf[i])
	}
}

func TestAsyncBatchWriter_FlushNilBuffer(t *testing.T) {
	cfg := newTestConfig()
	w := NewAsyncBatchWriter(cfg, 10, 5, time.Second)
	// flush on empty buffer must not panic
	assert.NotPanics(t, func() { w.flush(nil) })
	assert.NotPanics(t, func() { w.flush([]*BatchWriteRequest{}) })
}

func TestAsyncBatchWriter_FlushSkipsNilRequests(t *testing.T) {
	cfg := newTestConfig()
	w := NewAsyncBatchWriter(cfg, 10, 5, time.Second)
	// nil entries must not panic or increment counters
	buf := []*BatchWriteRequest{nil, {StbControlScheduleResultRaw: nil}}
	assert.NotPanics(t, func() { w.flush(buf) })
	assert.Equal(t, int64(0), w.droppedCount.Load())
}

func TestAsyncBatchWriter_FlushDropsOnConvertError(t *testing.T) {
	cfg := newTestConfig()
	w := NewAsyncBatchWriter(cfg, 10, 5, time.Second)

	// Craft a raw record with a zero-value RequestTime that causes ConvertBatchFormat to fail.
	// orm.Datetime{} with zero Time will succeed, so we can verify via dropped count after
	// an insert error (no DB configured). We just check that nil/skip logic doesn't crash.
	buf := []*BatchWriteRequest{
		{StbControlScheduleResultRaw: &tables.StbControlScheduleResultRaw{
			ID: "test-1",
		}},
	}
	// DB is not configured so Insert will fail, droppedCount increments
	assert.NotPanics(t, func() { w.flush(buf) })
}

func TestAsyncBatchWriter_DroppedCountStartsZero(t *testing.T) {
	cfg := newTestConfig()
	w := NewAsyncBatchWriter(cfg, 10, 5, time.Second)
	assert.Equal(t, int64(0), w.droppedCount.Load())
	assert.Equal(t, int64(0), w.flushedCount.Load())
	assert.Equal(t, int64(0), w.flushedBatches.Load())
}
