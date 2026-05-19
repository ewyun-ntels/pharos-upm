package gin

import (
	"bytes"
	"fmt"
	"log/slog"
	"strings"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSlogWriter_Write_ConcurrentAccess(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	logger := slog.New(slog.NewTextHandler(&buf, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}))

	writer := slogWriter{
		logger: logger,
		level:  slog.LevelInfo,
	}

	const numGoroutines = 10
	const messagesPerGoroutine = 5
	var wg sync.WaitGroup

	wg.Add(numGoroutines)

	for i := range numGoroutines {
		go func(id int) {
			defer wg.Done()

			for j := range messagesPerGoroutine {
				message := fmt.Sprintf("Test message from goroutine %d, iteration %d", id, j)
				n, err := writer.Write([]byte(message))

				assert.NoError(t, err, "Write should not return error")
				assert.Equal(t, len(message), n, "Should return correct number of bytes written")
			}
		}(i)
	}

	wg.Wait()

	// 버퍼에 로그가 기록되었는지 확인
	output := buf.String()
	assert.NotEmpty(t, output, "Buffer should contain log output")
}

func TestSlogWriter_WriteWithNewline_ConcurrentAccess(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	logger := slog.New(slog.NewTextHandler(&buf, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}))

	writer := slogWriter{
		logger: logger,
		level:  slog.LevelError,
	}

	const numGoroutines = 8
	var wg sync.WaitGroup
	results := make([]struct {
		bytesWritten int
		err          error
	}, numGoroutines)

	wg.Add(numGoroutines)

	for i := range numGoroutines {
		go func(id int) {
			defer wg.Done()

			// 개행 문자가 포함된 메시지
			message := fmt.Sprintf("Error message from goroutine %d\n", id)
			n, err := writer.Write([]byte(message))

			results[id].bytesWritten = n
			results[id].err = err
		}(i)
	}

	wg.Wait()

	// 모든 write 작업이 성공했는지 확인
	for i, result := range results {
		assert.NoError(t, result.err, "Goroutine %d should not have error", i)
		assert.Greater(t, result.bytesWritten, 0, "Goroutine %d should write bytes", i)
	}

	// 버퍼에 로그가 기록되었는지 확인
	output := buf.String()
	assert.NotEmpty(t, output, "Buffer should contain log output")
}

func TestSlogWriter_DifferentLevels_ConcurrentAccess(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	logger := slog.New(slog.NewTextHandler(&buf, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}))

	writers := []slogWriter{
		{logger: logger, level: slog.LevelDebug},
		{logger: logger, level: slog.LevelInfo},
		{logger: logger, level: slog.LevelWarn},
		{logger: logger, level: slog.LevelError},
	}

	const numGoroutines = 12
	var wg sync.WaitGroup

	wg.Add(numGoroutines)

	for i := range numGoroutines {
		go func(id int) {
			defer wg.Done()

			// 다양한 레벨의 writer 사용
			writer := writers[id%len(writers)]
			message := fmt.Sprintf("Message from goroutine %d with level %v", id, writer.level)

			n, err := writer.Write([]byte(message))

			assert.NoError(t, err, "Write should not return error for goroutine %d", id)
			assert.Equal(t, len(message), n, "Should return correct bytes for goroutine %d", id)
		}(i)
	}

	wg.Wait()

	// 버퍼에 모든 레벨의 로그가 기록되었는지 확인
	output := buf.String()
	assert.NotEmpty(t, output, "Buffer should contain log output")
}

func TestSlogWriter_EmptyMessage_ConcurrentAccess(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	logger := slog.New(slog.NewTextHandler(&buf, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}))

	writer := slogWriter{
		logger: logger,
		level:  slog.LevelInfo,
	}

	const numGoroutines = 6
	var wg sync.WaitGroup
	results := make([]int, numGoroutines)

	wg.Add(numGoroutines)

	for i := range numGoroutines {
		go func(id int) {
			defer wg.Done()

			// 빈 메시지 테스트
			n, err := writer.Write([]byte(""))

			assert.NoError(t, err, "Empty write should not return error")
			results[id] = n
		}(i)
	}

	wg.Wait()

	// 모든 빈 메시지 write가 0 바이트를 반환했는지 확인
	for i, n := range results {
		assert.Equal(t, 0, n, "Empty write from goroutine %d should return 0 bytes", i)
	}
}

func TestSlogWriter_LargeMessage_ConcurrentAccess(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	logger := slog.New(slog.NewTextHandler(&buf, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}))

	writer := slogWriter{
		logger: logger,
		level:  slog.LevelWarn,
	}

	// 큰 메시지 생성 (1KB)
	largeMessage := strings.Repeat("Large message content ", 50)

	const numGoroutines = 4
	var wg sync.WaitGroup
	results := make([]struct {
		bytesWritten int
		err          error
	}, numGoroutines)

	wg.Add(numGoroutines)

	for i := range numGoroutines {
		go func(id int) {
			defer wg.Done()

			message := fmt.Sprintf("Goroutine %d: %s", id, largeMessage)
			n, err := writer.Write([]byte(message))

			results[id].bytesWritten = n
			results[id].err = err
		}(i)
	}

	wg.Wait()

	// 모든 큰 메시지가 성공적으로 쓰였는지 확인
	for i, result := range results {
		assert.NoError(t, result.err, "Large message write from goroutine %d should not error", i)
		assert.Greater(t, result.bytesWritten, 1000, "Large message from goroutine %d should write many bytes", i)
	}

	// 버퍼에 내용이 있는지 확인
	output := buf.String()
	assert.NotEmpty(t, output, "Buffer should contain large message output")
}
