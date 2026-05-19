package util

import (
	"testing"

	"github.com/spf13/cobra"
)

func TestCommand_Creation(t *testing.T) {
	cmd := Command()

	if cmd == nil {
		t.Fatal("Command() returned nil")
	}

	// 기본 설정 확인
	expectedUse := "util"
	expectedShort := "utilities"

	if cmd.Use != expectedUse {
		t.Errorf("Use = %q, want %q", cmd.Use, expectedUse)
	}

	if cmd.Short != expectedShort {
		t.Errorf("Short = %q, want %q", cmd.Short, expectedShort)
	}

	if !cmd.SilenceUsage {
		t.Error("SilenceUsage should be true")
	}
}

func TestCommand_Subcommands(t *testing.T) {
	cmd := Command()

	expectedSubcommands := []string{
		"postgresql_metric_collect",
		"certification",
		"centrifuge_publish",
	}

	if !cmd.HasSubCommands() {
		t.Error("Command should have subcommands")
	}

	subcommands := cmd.Commands()
	if len(subcommands) != len(expectedSubcommands) {
		t.Errorf("Expected %d subcommands, got %d", len(expectedSubcommands), len(subcommands))
	}

	// 각 서브커맨드가 존재하는지 확인
	subcommandMap := make(map[string]*cobra.Command)
	for _, subcmd := range subcommands {
		subcommandMap[subcmd.Name()] = subcmd
	}

	for _, expected := range expectedSubcommands {
		if _, exists := subcommandMap[expected]; !exists {
			t.Errorf("Expected subcommand %q not found", expected)
		}
	}
}

func TestCommand_SubcommandExecution(t *testing.T) {
	cmd := Command()

	// 각 서브커맨드의 기본 구조 테스트
	tests := []struct {
		name       string
		subcommand string
		hasPreRunE bool
		hasRunE    bool
		hasRun     bool
	}{
		{
			name:       "postgresql_metric_collect",
			subcommand: "postgresql_metric_collect",
			hasPreRunE: true,
			hasRunE:    true,
		},
		{
			name:       "certification",
			subcommand: "certification",
			hasPreRunE: true,
			hasRunE:    true,
		},
		{
			name:       "centrifuge_publish",
			subcommand: "centrifuge_publish",
			hasPreRunE: true,
			hasRunE:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			subcmd, _, err := cmd.Find([]string{tt.subcommand})
			if err != nil {
				t.Fatalf("Failed to find subcommand %q: %v", tt.subcommand, err)
			}

			if tt.hasPreRunE && subcmd.PreRunE == nil {
				t.Errorf("Subcommand %q should have PreRunE", tt.subcommand)
			}

			if tt.hasRunE && subcmd.RunE == nil {
				t.Errorf("Subcommand %q should have RunE", tt.subcommand)
			}

			if tt.hasRun && subcmd.Run == nil {
				t.Errorf("Subcommand %q should have Run", tt.subcommand)
			}
		})
	}
}

func TestCommand_Help(t *testing.T) {
	cmd := Command()

	// 커맨드가 올바르게 생성되었는지 확인
	if cmd.Name() != "util" {
		t.Errorf("Expected command name 'util', got %q", cmd.Name())
	}
}

func TestCommand_Usage(t *testing.T) {
	cmd := Command()

	// Usage 문자열 생성이 에러 없이 되는지 확인
	usage := cmd.UsageString()
	if len(usage) == 0 {
		t.Error("Usage string should not be empty")
	}

	// 기본적인 내용이 포함되어 있는지 확인
	if !containsString(usage, "util") {
		t.Error("Usage should contain 'util'")
	}
}

// containsString는 문자열 포함 여부를 확인하는 헬퍼 함수
func containsString(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// 벤치마크 테스트
func BenchmarkCommand_Creation(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = Command()
	}
}

func BenchmarkCommand_Find(b *testing.B) {
	cmd := Command()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _, _ = cmd.Find([]string{"generate_hash"})
	}
}
