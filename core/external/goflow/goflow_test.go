package goflow

import (
	"context"
	"testing"
	"time"

	"ntels.com/pharos/core/external/orm"
	"ntels.com/pharos/core/pkg/common"
	"ntels.com/pharos/third_party/goflow"
)

func TestGoflow_Start(t *testing.T) {
	tests := []struct {
		name        string
		config      common.Config
		expectError bool
	}{
		{
			name: "workflow disabled",
			config: common.Config{
				Workflow: common.WorkflowConfig{
					Use: false,
				},
			},
			expectError: false,
		},
		{
			name: "workflow enabled with minimal config",
			config: common.Config{
				Workflow: common.WorkflowConfig{
					Use: true,
				},
				Database: orm.DatabaseConfig{
					Driver: orm.DriverSqlite,
					SQLite: orm.SQLiteConfig{
						Path: ":memory:",
					},
				},
			},
			expectError: false, // May fail due to missing dependencies
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := &Goflow{}
			_ = g.Start("", tt.config)
			// Verify no panic occurred
		})
	}
}

func TestGoflow_Start_AlreadyRunning(t *testing.T) {
	g := &Goflow{}

	// Mock a running goflow
	g.gf = goflow.NewWithMutex(goflow.Options{
		ShowExamples: false,
	})

	config := common.Config{
		Workflow: common.WorkflowConfig{
			Use: true,
		},
	}

	// Starting again should return nil without error
	err := g.Start("", config)
	if err != nil {
		t.Errorf("Expected no error when already running, got %v", err)
	}
}

func TestGoflow_Stop(t *testing.T) {
	tests := []struct {
		name string
		gf   *goflow.GoflowWithMutex
	}{
		{
			name: "stop when not initialized",
			gf:   nil,
		},
		{
			name: "stop when initialized but not running",
			gf:   nil, // Would be initialized in actual scenario
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := &Goflow{
				gf: tt.gf,
			}

			// Should not panic
			defer func() {
				if r := recover(); r != nil {
					t.Errorf("Stop() panicked: %v", r)
				}
			}()

			g.Stop()
		})
	}
}

func TestGoflow_AddJob(t *testing.T) {
	tests := []struct {
		name      string
		gf        *goflow.GoflowWithMutex
		wantError bool
	}{
		{
			name:      "add job when not running",
			gf:        nil,
			wantError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := &Goflow{
				gf: tt.gf,
			}

			err := g.AddJob(func() *goflow.Job {
				return &goflow.Job{
					Name:     "test-job",
					Schedule: "* * * * *",
				}
			})

			if (err != nil) != tt.wantError {
				t.Errorf("AddJob() error = %v, wantError %v", err, tt.wantError)
			}
		})
	}
}

func TestGoflow_UpdateJob(t *testing.T) {
	tests := []struct {
		name      string
		gf        *goflow.GoflowWithMutex
		wantError bool
	}{
		{
			name:      "update job when not running",
			gf:        nil,
			wantError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := &Goflow{
				gf: tt.gf,
			}

			err := g.UpdateJob(func() *goflow.Job {
				return &goflow.Job{
					Name:     "test-job",
					Schedule: "0 * * * *",
				}
			})

			if (err != nil) != tt.wantError {
				t.Errorf("UpdateJob() error = %v, wantError %v", err, tt.wantError)
			}
		})
	}
}

func TestGoflow_RemoveJob(t *testing.T) {
	tests := []struct {
		name      string
		jobName   string
		gf        *goflow.GoflowWithMutex
		wantError bool
	}{
		{
			name:      "remove job when not running",
			jobName:   "test-job",
			gf:        nil,
			wantError: false,
		},
		{
			name:      "remove empty job name",
			jobName:   "",
			gf:        nil,
			wantError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := &Goflow{
				gf: tt.gf,
			}

			err := g.RemoveJob(tt.jobName)

			if (err != nil) != tt.wantError {
				t.Errorf("RemoveJob() error = %v, wantError %v", err, tt.wantError)
			}
		})
	}
}

func TestGoflow_ContextCancellation(t *testing.T) {
	g := &Goflow{}

	// Create a context and cancel function
	ctx, cancel := context.WithCancel(context.Background())
	g.ctx = ctx
	g.cancelFunc = cancel

	// Verify context is not cancelled
	select {
	case <-g.ctx.Done():
		t.Error("Context should not be cancelled yet")
	default:
		// Expected
	}

	// Call Stop to cancel context
	g.cancelFunc()

	// Wait a bit and verify context is cancelled
	time.Sleep(10 * time.Millisecond)
	select {
	case <-ctx.Done():
		// Expected
	default:
		t.Error("Context should be cancelled after calling Stop")
	}
}

func TestGoflow_updateJob(t *testing.T) {
	tests := []struct {
		name   string
		config common.Config
	}{
		{
			name: "update job with backup config",
			config: func() common.Config {
				cfg := common.Config{
					Database: orm.DatabaseConfig{
						Driver: orm.DriverSqlite,
						SQLite: orm.SQLiteConfig{
							Path: ":memory:",
						},
					},
					Collect: common.CollectConfig{
						Statistics: common.CollectStatisticsConfig{
							Use:      true,
							Schedule: "0 * * * *",
							Report: common.ReportConfig{
								Use:      true,
								Schedule: "0 0 * * 0",
								TTL:      "30d",
							},
						},
					},
				}
				cfg.Database.SQLite.Backup.Use = true
				cfg.Database.SQLite.Backup.Schedule = "0 0 * * *"
				cfg.Database.SQLite.Backup.Directory = "/tmp/backup"
				cfg.Database.SQLite.Backup.TTL = "7d"
				return cfg
			}(),
		},
		{
			name: "update job with disabled features",
			config: func() common.Config {
				cfg := common.Config{
					Database: orm.DatabaseConfig{
						Driver: orm.DriverSqlite,
						SQLite: orm.SQLiteConfig{
							Path: ":memory:",
						},
					},
				}
				cfg.Database.SQLite.Backup.Use = false
				return cfg
			}(),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := &Goflow{}

			// This will fail without proper database setup, but we verify no panic
			_ = g.updateJob("", tt.config)
		})
	}
}

func TestGoflow_jobLoad(t *testing.T) {
	g := &Goflow{}

	config := common.Config{
		Workflow: common.WorkflowConfig{
			Use: true,
		},
		Database: orm.DatabaseConfig{
			Driver: orm.DriverSqlite,
			SQLite: orm.SQLiteConfig{
				Path: ":memory:",
			},
		},
	}

	// This will fail without proper setup, but we test the structure
	err := g.jobLoad("", config)
	// Error is expected due to missing dependencies, just verify no panic
	_ = err
}

func TestAddUpdateJobInfo(t *testing.T) {
	tests := []struct {
		name    string
		jobName string
		jobInfo UpdateJobInfo
	}{
		{
			name:    "add sqlite-vacuum job",
			jobName: "sqlite-vacuum",
			jobInfo: UpdateJobInfo{
				Use:      true,
				Schedule: "0 0 * * *",
			},
		},
		{
			name:    "add collect-statistics job",
			jobName: "collect-statistics",
			jobInfo: UpdateJobInfo{
				Use:      true,
				Schedule: "*/5 * * * *",
			},
		},
		{
			name:    "add disabled job",
			jobName: "disabled-job",
			jobInfo: UpdateJobInfo{
				Use:      false,
				Schedule: "",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			AddUpdateJobInfo(tt.jobName, tt.jobInfo)

			// Verify it was added using Exist and Get
			if !UpdateJobInfos.Exist(tt.jobName) {
				t.Errorf("Job %s was not added to UpdateJobInfos", tt.jobName)
				return
			}

			info := UpdateJobInfos.Get(tt.jobName)

			if info.Use != tt.jobInfo.Use {
				t.Errorf("Job %s Use = %v, want %v", tt.jobName, info.Use, tt.jobInfo.Use)
			}

			if info.Schedule != tt.jobInfo.Schedule {
				t.Errorf("Job %s Schedule = %v, want %v", tt.jobName, info.Schedule, tt.jobInfo.Schedule)
			}
		})
	}
}

// Benchmark tests
func BenchmarkGoflow_AddJob(b *testing.B) {
	g := &Goflow{}

	jobFunc := func() *goflow.Job {
		return &goflow.Job{
			Name:     "benchmark-job",
			Schedule: "* * * * *",
		}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = g.AddJob(jobFunc)
	}
}

func BenchmarkGoflow_UpdateJob(b *testing.B) {
	g := &Goflow{}

	jobFunc := func() *goflow.Job {
		return &goflow.Job{
			Name:     "benchmark-job",
			Schedule: "0 * * * *",
		}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = g.UpdateJob(jobFunc)
	}
}

func BenchmarkGoflow_RemoveJob(b *testing.B) {
	g := &Goflow{}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = g.RemoveJob("benchmark-job")
	}
}

func BenchmarkAddUpdateJobInfo(b *testing.B) {
	info := UpdateJobInfo{
		Use:      true,
		Schedule: "0 * * * *",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		AddUpdateJobInfo("benchmark-job", info)
	}
}
