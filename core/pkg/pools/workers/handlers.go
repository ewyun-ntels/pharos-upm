package workers

import (
	"context"
	"fmt"
	"log/slog"
	"time"
)

// SampleHandler1 is a sample handler for task type 1.
// This is sample code. Delete it and write your own handlers for actual projects.
//
// Handler implementation guide:
// 1. Add required configuration fields to the struct (DB connections, API clients, etc.)
// 2. Implement actual business logic in the Handle() method
// 3. Specify the JobType to handle in the CanHandle() method
type SampleHandler1 struct {
	// Example: Add configuration fields
	// dbConn *sql.DB
	// apiClient *http.Client
}

// NewSampleHandler1 creates a new sample handler
func NewSampleHandler1() *SampleHandler1 {
	return &SampleHandler1{}
}

// Handle processes a job - this is a sample implementation
func (h *SampleHandler1) Handle(ctx context.Context, job Job) error {
	// This is sample code. Replace with actual logic.

	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	// Actual implementation example:
	// 1. Extract data from Payload
	//    payload, ok := job.Payload.(YourPayloadType)
	//    if !ok { return fmt.Errorf("invalid payload type") }
	//
	// 2. Perform business logic
	//    result, err := h.processData(ctx, payload)
	//    if err != nil { return err }
	//
	// 3. Save result
	//    return h.saveResult(ctx, result)

	// Sample: Task simulation (improved pattern)
	timer := time.NewTimer(10 * time.Millisecond)
	defer timer.Stop()

	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return nil
	}
}

// CanHandle returns true if this handler can process the given job type
func (h *SampleHandler1) CanHandle(jobType JobType) bool {
	return jobType == JobTypeSampleTask1
}

// SampleHandler2 is a sample handler for task type 2.
// This is sample code. Delete it and write your own handlers for actual projects.
type SampleHandler2 struct {
	// Example: Add configuration fields
}

// NewSampleHandler2 creates a new sample handler
func NewSampleHandler2() *SampleHandler2 {
	return &SampleHandler2{}
}

// Handle processes a job - this is a sample implementation
func (h *SampleHandler2) Handle(ctx context.Context, job Job) error {
	// This is sample code. Replace with actual logic.

	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	// Write business logic here for actual implementation

	// Sample: Task simulation (improved pattern)
	timer := time.NewTimer(5 * time.Millisecond)
	defer timer.Stop()

	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return nil
	}
}

// CanHandle returns true if this handler can process the given job type
func (h *SampleHandler2) CanHandle(jobType JobType) bool {
	return jobType == JobTypeSampleTask2
}

// BaseHandler provides a base implementation for easily creating custom handlers.
// Use this to quickly create simple handlers.
//
// Usage example:
//
//	customType := workers.JobType("analysis")
//	handler := workers.NewBaseHandler(customType, func(ctx context.Context, job workers.Job) error {
//	    // Actual processing logic
//	    fmt.Printf("Processing job: %s\n", job.ID)
//	    return nil
//	})
//	pool.RegisterHandler(handler)
type BaseHandler struct {
	jobType    JobType
	handleFunc func(context.Context, Job) error
}

// NewBaseHandler creates a new base handler
func NewBaseHandler(jobType JobType, handleFunc func(context.Context, Job) error) *BaseHandler {
	return &BaseHandler{
		jobType:    jobType,
		handleFunc: handleFunc,
	}
}

// Handle processes a job using the provided handle function
func (h *BaseHandler) Handle(ctx context.Context, job Job) error {
	if h.handleFunc == nil {
		return fmt.Errorf("handle function not set")
	}
	return h.handleFunc(ctx, job)
}

// CanHandle returns true if this handler can process the given job type
func (h *BaseHandler) CanHandle(jobType JobType) bool {
	return h.jobType == jobType
}

// HandlerFunc is a function type that implements JobHandler interface
type HandlerFunc func(ctx context.Context, job Job) error

// Handle implements JobHandler.Handle
func (f HandlerFunc) Handle(ctx context.Context, job Job) error {
	return f(ctx, job)
}

// CanHandle implements JobHandler.CanHandle (returns true for all types)
func (f HandlerFunc) CanHandle(jobType JobType) bool {
	return true // HandlerFunc is generic
}

// Middleware wraps a JobHandler with additional functionality
// Follows the decorator pattern for composable behavior
//
// Common use cases:
//   - Logging: Track handler execution and performance
//   - Timeout: Enforce execution time limits
//   - Recovery: Prevent panics from crashing workers
//   - Authentication: Validate job permissions
//   - Rate limiting: Control handler execution rate
//
// Middleware can be chained using the Chain function:
//
//	handler := Chain(
//	    WithLogging,
//	    WithTimeout(5*time.Second),
//	    WithRecovery,
//	)(baseHandler)
type Middleware func(JobHandler) JobHandler

// WithLogging is a middleware that adds structured logging to a handler
// Logs job execution start, completion, and any errors with performance metrics
//
// Logged fields:
//   - job_id: Unique identifier for the job
//   - job_type: Type/category of the job
//   - duration_ms: Execution time in milliseconds
//   - error: Error message if handler failed (only on error)
//
// Example output:
//
//	INFO Job completed job_id=abc-123 job_type=email duration_ms=45.23
//	ERROR Job processing failed job_id=def-456 job_type=report duration_ms=102.50 error="db timeout"
func WithLogging(next JobHandler) JobHandler {
	return HandlerFunc(func(ctx context.Context, job Job) error {
		start := time.Now()
		err := next.Handle(ctx, job)
		duration := time.Since(start)
		durationMs := float64(duration.Nanoseconds()) / 1e6

		if err != nil {
			slog.Error("Job processing failed",
				"job_id", job.ID,
				"job_type", job.Type,
				"duration_ms", fmt.Sprintf("%.2f", durationMs),
				"error", err.Error())
		} else {
			slog.Info("Job completed",
				"job_id", job.ID,
				"job_type", job.Type,
				"duration_ms", fmt.Sprintf("%.2f", durationMs))
		}
		return err
	})
}

// WithTimeout is a middleware that adds timeout enforcement to a handler
// Returns an error if handler execution exceeds the specified duration
//
// The timeout applies to the handler execution only, not to the entire job lifecycle
// If the handler respects context cancellation, it will be interrupted cleanly
//
// Usage:
//
//	timeoutHandler := WithTimeout(5*time.Second)(baseHandler)
//	pool.RegisterHandler(timeoutHandler)
//
// Note: The wrapped handler should check ctx.Done() periodically for clean cancellation
func WithTimeout(timeout time.Duration) Middleware {
	return func(next JobHandler) JobHandler {
		return HandlerFunc(func(ctx context.Context, job Job) error {
			ctx, cancel := context.WithTimeout(ctx, timeout)
			defer cancel()

			done := make(chan error, 1)
			go func() {
				done <- next.Handle(ctx, job)
			}()

			select {
			case err := <-done:
				return err
			case <-ctx.Done():
				timeoutMs := float64(timeout.Nanoseconds()) / 1e6
				return fmt.Errorf("job %s timed out after %.0fms", job.ID, timeoutMs)
			}
		})
	}
}

// WithRecovery is a middleware that recovers from panics in handler execution
// Prevents a single handler panic from crashing the entire worker pool
//
// Recovery behavior:
//   - Catches any panic from the handler
//   - Converts panic to error with panic value in error message
//   - Returns error instead of propagating panic
//   - Job will be marked as failed and may be retried
//
// This middleware should typically be the outermost layer in the middleware chain
// to ensure all panics are caught, even from other middleware:
//
//	handler := Chain(WithLogging, WithRecovery)(baseHandler)
//
// Note: For production use, consider adding stack trace logging for debugging
func WithRecovery(next JobHandler) JobHandler {
	return HandlerFunc(func(ctx context.Context, job Job) (err error) {
		defer func() {
			if r := recover(); r != nil {
				err = fmt.Errorf("job %s panicked: %v", job.ID, r)
			}
		}()
		return next.Handle(ctx, job)
	})
}

// Chain combines multiple middlewares into one
// Middleware are applied from left to right (first middleware is outermost)
//
// Example:
//
//	combined := Chain(
//	    WithRecovery,         // Applied first (outermost)
//	    WithTimeout(5*time.Second), // Applied second
//	    WithLogging,          // Applied last (innermost)
//	)
//	handler := combined(baseHandler)
//
// Execution order:
//
//	WithRecovery -> WithTimeout -> WithLogging -> baseHandler
//
// This is useful for creating reusable middleware stacks:
//
//	standardMiddleware := Chain(WithRecovery, WithLogging)
//	pool.RegisterHandler(standardMiddleware(emailHandler))
//	pool.RegisterHandler(standardMiddleware(reportHandler))
func Chain(middlewares ...Middleware) Middleware {
	return func(next JobHandler) JobHandler {
		for i := len(middlewares) - 1; i >= 0; i-- {
			next = middlewares[i](next)
		}
		return next
	}
}

// ApplyMiddleware applies a middleware to a handler
func ApplyMiddleware(handler JobHandler, middleware Middleware) JobHandler {
	return middleware(handler)
}
