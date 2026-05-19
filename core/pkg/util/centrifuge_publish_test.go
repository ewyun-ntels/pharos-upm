package util

import (
	"errors"
	"testing"

	"github.com/spf13/cobra"
)

func TestCentrifugePublishCommand_ValidArgs(t *testing.T) {
	tests := []struct {
		name string
		args []string
		want error
	}{
		{
			name: "valid args",
			args: []string{"ws://localhost:8000/connection/websocket", "test-channel", "test-data"},
			want: nil,
		},
		{
			name: "insufficient args",
			args: []string{"ws://localhost:8000/connection/websocket", "test-channel"},
			want: errors.New("invalid args([ws://localhost:8000/connection/websocket test-channel])"),
		},
		{
			name: "too many args",
			args: []string{"ws://localhost:8000/connection/websocket", "test-channel", "test-data", "extra"},
			want: errors.New("invalid args([ws://localhost:8000/connection/websocket test-channel test-data extra])"),
		},
		{
			name: "empty args",
			args: []string{},
			want: errors.New("invalid args([])"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := &cobra.Command{}
			err := centrifugePublish.PreRunE(cmd, tt.args)

			if tt.want == nil {
				if err != nil {
					t.Errorf("PreRunE() error = %v, want nil", err)
				}
			} else {
				if err == nil {
					t.Errorf("PreRunE() error = nil, want %v", tt.want)
				} else if err.Error() != tt.want.Error() {
					t.Errorf("PreRunE() error = %v, want %v", err, tt.want)
				}
			}
		})
	}
}

func TestCentrifugePublishCommand_Usage(t *testing.T) {
	expectedUse := "centrifuge_publish endpoint channel data"
	expectedShort := "centrifuge publish"

	if centrifugePublish.Use != expectedUse {
		t.Errorf("Use = %q, want %q", centrifugePublish.Use, expectedUse)
	}

	if centrifugePublish.Short != expectedShort {
		t.Errorf("Short = %q, want %q", centrifugePublish.Short, expectedShort)
	}
}

// 벤치마크 테스트
func BenchmarkCentrifugePublishCommand_PreRunE(b *testing.B) {
	args := []string{"ws://localhost:8000/connection/websocket", "test-channel", "test-data"}
	cmd := &cobra.Command{}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = centrifugePublish.PreRunE(cmd, args)
	}
}

// 파라미터 검증 테스트
func TestCentrifugePublishCommand_ParameterValidation(t *testing.T) {
	testCases := []struct {
		name     string
		endpoint string
		channel  string
		data     string
		wantErr  bool
	}{
		{
			name:     "valid parameters",
			endpoint: "ws://localhost:8000/connection/websocket",
			channel:  "test-channel",
			data:     "test-data",
			wantErr:  false,
		},
		{
			name:     "empty endpoint",
			endpoint: "",
			channel:  "test-channel",
			data:     "test-data",
			wantErr:  false, // PreRunE는 빈 값에 대한 검증을 하지 않음
		},
		{
			name:     "empty channel",
			endpoint: "ws://localhost:8000/connection/websocket",
			channel:  "",
			data:     "test-data",
			wantErr:  false, // PreRunE는 빈 값에 대한 검증을 하지 않음
		},
		{
			name:     "empty data",
			endpoint: "ws://localhost:8000/connection/websocket",
			channel:  "test-channel",
			data:     "",
			wantErr:  false, // PreRunE는 빈 값에 대한 검증을 하지 않음
		},
		{
			name:     "special characters in channel",
			endpoint: "ws://localhost:8000/connection/websocket",
			channel:  "test-channel-한글-特殊字符",
			data:     "test-data",
			wantErr:  false,
		},
		{
			name:     "json data",
			endpoint: "ws://localhost:8000/connection/websocket",
			channel:  "test-channel",
			data:     `{"message": "hello", "timestamp": 1234567890}`,
			wantErr:  false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			args := []string{tc.endpoint, tc.channel, tc.data}
			cmd := &cobra.Command{}

			err := centrifugePublish.PreRunE(cmd, args)

			if tc.wantErr {
				if err == nil {
					t.Errorf("PreRunE() error = nil, want error")
				}
			} else {
				if err != nil {
					t.Errorf("PreRunE() error = %v, want nil", err)
				}
			}
		})
	}
}

// Example 테스트
func ExampleCentrifugePublish() {
	// 이 예제는 실제 Centrifuge 서버가 실행 중일 때 작동합니다
	cmd := &cobra.Command{}
	args := []string{"ws://localhost:8000/connection/websocket", "notifications", "Hello World!"}

	// 매개변수 검증
	if err := centrifugePublish.PreRunE(cmd, args); err != nil {
		panic(err)
	}

	// 실제 실행은 서버가 필요하므로 주석 처리
	// if err := centrifugePublish.RunE(cmd, args); err != nil {
	//     panic(err)
	// }

	// Output:
	//
}
