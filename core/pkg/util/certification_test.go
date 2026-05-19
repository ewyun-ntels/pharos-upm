package util

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/spf13/cobra"
	coreV1 "k8s.io/api/core/v1"
	"ntels.com/pharos/core/internal/cert"
)

func TestCertificationCommand_ValidArgs(t *testing.T) {
	tests := []struct {
		name string
		args []string
		want error
	}{
		{
			name: "valid args",
			args: []string{"localhost,example.com", "127.0.0.1,192.168.1.1", "2025-12-31", "/tmp/test-cert"},
			want: nil,
		},
		{
			name: "insufficient args",
			args: []string{"localhost", "127.0.0.1", "2025-12-31"},
			want: fmt.Errorf("invalid args([localhost 127.0.0.1 2025-12-31])"),
		},
		{
			name: "too many args",
			args: []string{"localhost", "127.0.0.1", "2025-12-31", "/tmp/cert", "extra"},
			want: fmt.Errorf("invalid args([localhost 127.0.0.1 2025-12-31 /tmp/cert extra])"),
		},
		{
			name: "invalid date format",
			args: []string{"localhost", "127.0.0.1", "invalid-date", "/tmp/cert"},
			want: fmt.Errorf(`parsing time "invalid-date" as "2006-01-02": cannot parse "invalid-date" as "2006"`),
		},
		{
			name: "empty args",
			args: []string{},
			want: fmt.Errorf("invalid args([])"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := &cobra.Command{}
			err := certification.PreRunE(cmd, tt.args)

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

func TestCertificationCommand_RunE(t *testing.T) {
	// 테스트용 임시 디렉토리 생성
	tempDir := t.TempDir()

	tests := []struct {
		name    string
		args    []string
		wantErr bool
	}{
		{
			name:    "successful certificate generation",
			args:    []string{"localhost,example.com", "127.0.0.1,192.168.1.1", "2025-12-31", tempDir},
			wantErr: false,
		},
		{
			name:    "empty dns names",
			args:    []string{"", "127.0.0.1", "2025-12-31", tempDir},
			wantErr: false,
		},
		{
			name:    "empty ip addresses",
			args:    []string{"localhost", "", "2025-12-31", tempDir},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 각 테스트를 위한 서브디렉토리 생성
			testDir := filepath.Join(tempDir, tt.name)
			tt.args[3] = testDir

			cmd := &cobra.Command{}
			err := certification.RunE(cmd, tt.args)

			if tt.wantErr {
				if err == nil {
					t.Errorf("RunE() error = nil, want error")
				}
			} else {
				if err != nil {
					t.Errorf("RunE() error = %v, want nil", err)
				}

				// 인증서 파일들이 생성되었는지 확인
				checkCertificateFiles(t, testDir)
			}
		})
	}
}

func checkCertificateFiles(t *testing.T, dir string) {
	expectedFiles := []string{
		coreV1.ServiceAccountRootCAKey, // ca.crt
		coreV1.TLSCertKey,              // tls.crt
		coreV1.TLSPrivateKeyKey,        // tls.key
	}

	for _, filename := range expectedFiles {
		filePath := filepath.Join(dir, filename)
		if _, err := os.Stat(filePath); err != nil {
			t.Errorf("Expected certificate file %s not found: %v", filename, err)
		} else {
			// 파일이 존재하면 내용이 있는지 확인
			info, err := os.Stat(filePath)
			if err != nil {
				t.Errorf("Error getting file info for %s: %v", filename, err)
			} else if info.Size() == 0 {
				t.Errorf("Certificate file %s is empty", filename)
			}

			// 파일 권한 확인 (0600이어야 함)
			expectedPerm := fs.FileMode(0600)
			if info.Mode().Perm() != expectedPerm {
				t.Errorf("File %s has incorrect permissions %v, want %v", filename, info.Mode().Perm(), expectedPerm)
			}
		}
	}
}

func TestCertificationCommand_Usage(t *testing.T) {
	expectedUse := "certification subject-alternate-name-dns-names subject-alternate-name-ips not-after destination"
	expectedShort := "certification"

	if certification.Use != expectedUse {
		t.Errorf("Use = %q, want %q", certification.Use, expectedUse)
	}

	if certification.Short != expectedShort {
		t.Errorf("Short = %q, want %q", certification.Short, expectedShort)
	}
}

func TestGeneratePem_Function(t *testing.T) {
	commonName := "test-pharos"
	organization := []string{"test-org"}
	dnsNames := []string{"localhost", "example.com"}
	ipAddresses := []string{"127.0.0.1", "192.168.1.1"}
	notAfter := time.Now().Add(365 * 24 * time.Hour) // 1년 후

	caCertPEM, certPEM, keyPEM, err := cert.GeneratePem(commonName, organization, dnsNames, ipAddresses, notAfter)

	if err != nil {
		t.Fatalf("GeneratePem() error = %v", err)
	}

	// PEM 버퍼들이 nil이 아닌지 확인
	if caCertPEM == nil {
		t.Error("caCertPEM is nil")
	}
	if certPEM == nil {
		t.Error("certPEM is nil")
	}
	if keyPEM == nil {
		t.Error("keyPEM is nil")
	}

	// PEM 버퍼들이 비어있지 않은지 확인
	if caCertPEM.Len() == 0 {
		t.Error("caCertPEM is empty")
	}
	if certPEM.Len() == 0 {
		t.Error("certPEM is empty")
	}
	if keyPEM.Len() == 0 {
		t.Error("keyPEM is empty")
	}

	// PEM 헤더 확인
	caCertStr := caCertPEM.String()
	if !contains(caCertStr, "-----BEGIN CERTIFICATE-----") || !contains(caCertStr, "-----END CERTIFICATE-----") {
		t.Error("caCertPEM doesn't contain valid PEM certificate headers")
	}

	certStr := certPEM.String()
	if !contains(certStr, "-----BEGIN CERTIFICATE-----") || !contains(certStr, "-----END CERTIFICATE-----") {
		t.Error("certPEM doesn't contain valid PEM certificate headers")
	}

	keyStr := keyPEM.String()
	if !contains(keyStr, "-----BEGIN RSA PRIVATE KEY-----") || !contains(keyStr, "-----END RSA PRIVATE KEY-----") {
		t.Error("keyPEM doesn't contain valid PEM private key headers")
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) &&
		len(substr) > 0 &&
		s[indexOf(s, substr):indexOf(s, substr)+len(substr)] == substr
}

func indexOf(s, substr string) int {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}

func TestCertificationCommand_DateParsing(t *testing.T) {
	tests := []struct {
		name    string
		dateStr string
		wantErr bool
	}{
		{
			name:    "valid date",
			dateStr: "2025-12-31",
			wantErr: false,
		},
		{
			name:    "valid date with leading zeros",
			dateStr: "2025-01-01",
			wantErr: false,
		},
		{
			name:    "invalid format",
			dateStr: "12-31-2025",
			wantErr: true,
		},
		{
			name:    "incomplete date",
			dateStr: "2025-12",
			wantErr: true,
		},
		{
			name:    "invalid month",
			dateStr: "2025-13-31",
			wantErr: true,
		},
		{
			name:    "invalid day",
			dateStr: "2025-12-32",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := time.Parse(time.DateOnly, tt.dateStr)

			if tt.wantErr {
				if err == nil {
					t.Errorf("time.Parse() error = nil, want error for date %s", tt.dateStr)
				}
			} else {
				if err != nil {
					t.Errorf("time.Parse() error = %v, want nil for date %s", err, tt.dateStr)
				}
			}
		})
	}
}

func TestCertificationCommand_DirectoryCreation(t *testing.T) {
	tempDir := t.TempDir()
	nestedDir := filepath.Join(tempDir, "nested", "test", "directory")

	args := []string{"localhost", "127.0.0.1", "2025-12-31", nestedDir}

	cmd := &cobra.Command{}
	err := certification.RunE(cmd, args)

	if err != nil {
		t.Fatalf("RunE() error = %v", err)
	}

	// 중첩된 디렉토리가 생성되었는지 확인
	if _, err := os.Stat(nestedDir); err != nil {
		t.Errorf("Nested directory not created: %v", err)
	}

	// 디렉토리 권한 확인 (0700이어야 함)
	info, err := os.Stat(nestedDir)
	if err != nil {
		t.Errorf("Error getting directory info: %v", err)
	} else {
		expectedPerm := fs.FileMode(0700)
		if info.Mode().Perm() != expectedPerm {
			t.Errorf("Directory has incorrect permissions %v, want %v", info.Mode().Perm(), expectedPerm)
		}
	}
}

// 벤치마크 테스트
func BenchmarkGeneratePem(b *testing.B) {
	commonName := "benchmark-test"
	organization := []string{"benchmark-org"}
	dnsNames := []string{"localhost", "example.com"}
	ipAddresses := []string{"127.0.0.1"}
	notAfter := time.Now().Add(365 * 24 * time.Hour)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _, _, err := cert.GeneratePem(commonName, organization, dnsNames, ipAddresses, notAfter)
		if err != nil {
			b.Fatalf("GeneratePem() error = %v", err)
		}
	}
}
