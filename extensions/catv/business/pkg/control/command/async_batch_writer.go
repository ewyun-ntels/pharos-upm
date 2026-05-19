package command

import (
	"log/slog"
	"sync/atomic"
	"time"

	"ntels.com/pharos/core/external/clickhouse_interface"
	"ntels.com/pharos/core/pkg/common"
	"ntels.com/pharos/extensions/catv/business/pkg/control/common/tables"
)

// BatchWriteRequest represents a batch write request for stb_control_job table.
type BatchWriteRequest struct {
	StbControlScheduleResultRaw *tables.StbControlScheduleResultRaw
}

// AsyncBatchWriter handles asynchronous batch writes to ClickHouse.
// It collects writes from a channel and flushes them in batches to minimize DB round-trips.
//
// Stop() must be called only after all senders have finished writing to the channel
// (e.g. after workerPool.Stop()), otherwise items in-flight may be missed by the drain loop.
type AsyncBatchWriter struct {
	config         common.Config
	resultChan     chan *BatchWriteRequest
	tableName      string // Resolved once at construction to avoid per-flush allocation
	batchSize      int
	flushInterval  time.Duration
	done           chan struct{}
	stopped        chan struct{} // Signals when goroutine has fully stopped
	flushedCount   atomic.Int64  // Total number of records successfully flushed
	flushedBatches atomic.Int64  // Total number of batch flushes
	droppedCount   atomic.Int64  // Total number of records dropped due to format/add errors
}

// NewAsyncBatchWriter creates a new AsyncBatchWriter instance.
func NewAsyncBatchWriter(config common.Config, channelSize, batchSize int, flushInterval time.Duration) *AsyncBatchWriter {
	return &AsyncBatchWriter{
		config:        config,
		resultChan:    make(chan *BatchWriteRequest, channelSize),
		tableName:     tables.NewStbControlScheduleResultTable(config).GetDistributedTableName(),
		batchSize:     batchSize,
		flushInterval: flushInterval,
		done:          make(chan struct{}),
		stopped:       make(chan struct{}),
	}
}

// Start begins the async batch writer goroutine.
func (w *AsyncBatchWriter) Start() {
	go func() {
		defer close(w.stopped) // Signal completion when goroutine exits
		ticker := time.NewTicker(w.flushInterval)
		defer ticker.Stop()

		buffer := make([]*BatchWriteRequest, 0, w.batchSize)

		for {
			select {
			case req := <-w.resultChan:
				buffer = append(buffer, req)
				if len(buffer) >= w.batchSize {
					w.flush(buffer)
					buffer = clearBuffer(buffer)
				}
			case <-ticker.C:
				if len(buffer) > 0 {
					w.flush(buffer)
					buffer = clearBuffer(buffer)
				}
			case <-w.done:
				// Flush remaining buffer before draining.
				// Precondition: all senders must have finished before Stop() is called.
				if len(buffer) > 0 {
					w.flush(buffer)
					buffer = clearBuffer(buffer)
				}
				// Drain any items that arrived between the last receive and now.
				for {
					select {
					case req := <-w.resultChan:
						buffer = append(buffer, req)
						if len(buffer) >= w.batchSize {
							w.flush(buffer)
							buffer = clearBuffer(buffer)
						}
					default:
						if len(buffer) > 0 {
							w.flush(buffer)
						}
						slog.Info("Async batch writer stopped",
							"flushed_records", w.flushedCount.Load(),
							"flushed_batches", w.flushedBatches.Load(),
							"dropped_records", w.droppedCount.Load())
						return
					}
				}
			}
		}
	}()

	slog.Info("Async batch writer started",
		"channel_size", cap(w.resultChan),
		"batch_size", w.batchSize,
		"flush_interval", w.flushInterval.Seconds())
}

// Stop signals the async batch writer to stop and flush remaining data.
// This method closes the done channel to trigger graceful shutdown and waits
// for the goroutine to finish, ensuring all pending writes are completed.
func (w *AsyncBatchWriter) Stop() {
	close(w.done)
	<-w.stopped // Wait for goroutine to finish
}

// flush writes a batch of records to ClickHouse.
// Records that fail format conversion or batch serialization are dropped and counted in droppedCount.
// Insert failures are logged but not retried — the batch is discarded.
func (w *AsyncBatchWriter) flush(buffer []*BatchWriteRequest) {
	if len(buffer) == 0 {
		return
	}

	batch := clickhouse_interface.ClickhouseBatch{}
	var recordCount int64

	for _, req := range buffer {
		if req == nil || req.StbControlScheduleResultRaw == nil {
			continue
		}
		data, err := req.StbControlScheduleResultRaw.ConvertBatchFormat()
		if err != nil {
			slog.Error("Failed to convert batch format, record dropped",
				"result_id", req.StbControlScheduleResultRaw.ID,
				"k8s_job_name", req.StbControlScheduleResultRaw.K8sJobName,
				"cm_mac_addr", req.StbControlScheduleResultRaw.CmMacAddr,
				"error", err)
			w.droppedCount.Add(1)
			continue
		}
		if err := batch.AddBatchJson(data); err != nil {
			slog.Error("Failed to add record to batch, record dropped",
				"result_id", req.StbControlScheduleResultRaw.ID,
				"k8s_job_name", req.StbControlScheduleResultRaw.K8sJobName,
				"cm_mac_addr", req.StbControlScheduleResultRaw.CmMacAddr,
				"error", err)
			w.droppedCount.Add(1)
		} else {
			recordCount++
		}
	}

	if recordCount == 0 {
		return
	}

	// NOTE: Insert failures are not retried. On transient ClickHouse errors the
	// entire batch is lost. Callers that require stronger guarantees should
	// implement retry at a higher level or rely on the 'running' status recovery
	// path in SelectWhereK8sJobName.
	// NOTE: recordCount is captured before Insert because ClickhouseBatch.Insert()
	// internally calls Get() which clears batchData, making Len() return 0 afterward.
	if err := batch.Insert(w.config.Catv.Database, w.tableName); err != nil {
		slog.Error("Failed to flush batch, batch dropped",
			"count", recordCount,
			"table", w.tableName,
			"error", err)
		w.droppedCount.Add(recordCount)
	} else {
		w.flushedCount.Add(recordCount)
		w.flushedBatches.Add(1)
		slog.Debug("Flushed batch",
			"count", recordCount,
			"table", w.tableName,
			"total_flushed", w.flushedCount.Load(),
			"total_batches", w.flushedBatches.Load())
	}
}

// GetChannel returns the send-only result channel for write requests.
func (w *AsyncBatchWriter) GetChannel() chan<- *BatchWriteRequest {
	return w.resultChan
}

// clearBuffer releases pointer references in the buffer to allow GC and resets its length.
func clearBuffer(buf []*BatchWriteRequest) []*BatchWriteRequest {
	for i := range buf {
		buf[i] = nil
	}
	return buf[:0]
}
