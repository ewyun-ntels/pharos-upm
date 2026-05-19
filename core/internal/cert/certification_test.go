package cert

import (
	"crypto/tls"
	"os"
	"path/filepath"
	"testing"
	"time"

	"ntels.com/pharos/core/pkg/common"
)

func TestParseTLSVersion(t *testing.T) {
	t.Parallel() // 병렬 실행

	tests := []struct {
		name       string
		version    string
		wantVer    uint16
		wantParsed bool
	}{
		{
			name:       "empty string defaults to TLS 1.3",
			version:    "",
			wantVer:    tls.VersionTLS13,
			wantParsed: false,
		},
		{
			name:       "default keyword",
			version:    "default",
			wantVer:    tls.VersionTLS13,
			wantParsed: false,
		},
		{
			name:       "TLS 1.2 numeric",
			version:    "1.2",
			wantVer:    tls.VersionTLS12,
			wantParsed: true,
		},
		{
			name:       "TLS 1.2 uppercase",
			version:    "TLS1.2",
			wantVer:    tls.VersionTLS12,
			wantParsed: true,
		},
		{
			name:       "TLS 1.2 lowercase",
			version:    "tls1.2",
			wantVer:    tls.VersionTLS12,
			wantParsed: true,
		},
		{
			name:       "TLS 1.3 numeric",
			version:    "1.3",
			wantVer:    tls.VersionTLS13,
			wantParsed: true,
		},
		{
			name:       "TLS 1.3 uppercase",
			version:    "TLS1.3",
			wantVer:    tls.VersionTLS13,
			wantParsed: true,
		},
		{
			name:       "TLS 1.3 lowercase",
			version:    "tls1.3",
			wantVer:    tls.VersionTLS13,
			wantParsed: true,
		},
		{
			name:       "invalid version",
			version:    "1.1",
			wantVer:    0,
			wantParsed: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel() // 서브테스트도 병렬 실행
			gotVer, gotParsed := parseTLSVersion(tt.version)
			if gotVer != tt.wantVer {
				t.Errorf("parseTLSVersion() version = %v, want %v", gotVer, tt.wantVer)
			}
			if gotParsed != tt.wantParsed {
				t.Errorf("parseTLSVersion() parsed = %v, want %v", gotParsed, tt.wantParsed)
			}
		})
	}
}

func TestResolveTLSVersions(t *testing.T) {
	t.Parallel() // 병렬 실행

	tests := []struct {
		name    string
		config  common.ServerConfig
		wantMin uint16
		wantMax uint16
	}{
		{
			name: "default versions (empty)",
			config: common.ServerConfig{
				TLSMinVersion: "",
				TLSMaxVersion: "",
			},
			wantMin: tls.VersionTLS13,
			wantMax: tls.VersionTLS13,
		},
		{
			name: "TLS 1.2 to 1.3",
			config: common.ServerConfig{
				TLSMinVersion: "1.2",
				TLSMaxVersion: "1.3",
			},
			wantMin: tls.VersionTLS12,
			wantMax: tls.VersionTLS13,
		},
		{
			name: "TLS 1.2 only",
			config: common.ServerConfig{
				TLSMinVersion: "1.2",
				TLSMaxVersion: "1.2",
			},
			wantMin: tls.VersionTLS12,
			wantMax: tls.VersionTLS12,
		},
		{
			name: "invalid versions fall back to default",
			config: common.ServerConfig{
				TLSMinVersion: "invalid",
				TLSMaxVersion: "invalid",
			},
			wantMin: tls.VersionTLS13,
			wantMax: tls.VersionTLS13,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel() // 서브테스트도 병렬 실행
			gotMin, gotMax := resolveTLSVersions(tt.config)
			if gotMin != tt.wantMin {
				t.Errorf("resolveTLSVersions() min = %v, want %v", gotMin, tt.wantMin)
			}
			if gotMax != tt.wantMax {
				t.Errorf("resolveTLSVersions() max = %v, want %v", gotMax, tt.wantMax)
			}
		})
	}
}

func TestSafeReadFile(t *testing.T) {
	t.Parallel() // 병렬 실행

	// Create a temporary directory for testing
	tmpDir := t.TempDir()

	// Create a test file
	testFile := filepath.Join(tmpDir, "test.txt")
	testContent := []byte("test content")
	if err := os.WriteFile(testFile, testContent, 0o600); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	tests := []struct {
		name      string
		baseDir   string
		filePath  string
		wantError bool
	}{
		{
			name:      "valid file read",
			baseDir:   tmpDir,
			filePath:  testFile,
			wantError: false,
		},
		{
			name:      "file outside base directory",
			baseDir:   tmpDir,
			filePath:  "/etc/passwd",
			wantError: true,
		},
		{
			name:      "non-existent file",
			baseDir:   tmpDir,
			filePath:  filepath.Join(tmpDir, "nonexistent.txt"),
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel() // 서브테스트도 병렬 실행
			data, err := safeReadFile(tt.baseDir, tt.filePath)
			if (err != nil) != tt.wantError {
				t.Errorf("safeReadFile() error = %v, wantError %v", err, tt.wantError)
				return
			}
			if !tt.wantError && string(data) != string(testContent) {
				t.Errorf("safeReadFile() = %v, want %v", string(data), string(testContent))
			}
		})
	}
}

func TestSetCertification_WithExplicitFiles(t *testing.T) {
	t.Parallel() // 병렬 실행

	// Create temporary directory and certificate files
	tmpDir := t.TempDir()

	// Generate test certificates (빠른 버전 사용)
	notAfter := time.Now().Add(24 * time.Hour)
	_, certPEM, keyPEM, err := generateTestPem(
		"test-common-name",
		[]string{"test-org"},
		[]string{"localhost"},
		[]string{"127.0.0.1"},
		notAfter,
	)
	if err != nil {
		t.Fatalf("Failed to generate test certificates: %v", err)
	}

	// Write certificate files
	certFile := filepath.Join(tmpDir, "test.crt")
	keyFile := filepath.Join(tmpDir, "test.key")

	if err := os.WriteFile(certFile, certPEM.Bytes(), 0o600); err != nil {
		t.Fatalf("Failed to write cert file: %v", err)
	}
	if err := os.WriteFile(keyFile, keyPEM.Bytes(), 0o600); err != nil {
		t.Fatalf("Failed to write key file: %v", err)
	}

	config := common.ServerConfig{
		Port: 8443,
		Cert: common.ServerCertConfig{
			CertFile: certFile,
			KeyFile:  keyFile,
		},
		TLSMinVersion: "1.2",
		TLSMaxVersion: "1.3",
	}

	tlsConfig, err := SetCertification(config)
	if err != nil {
		t.Fatalf("SetCertification() error = %v", err)
	}

	if tlsConfig == nil {
		t.Fatal("SetCertification() returned nil config")
	}

	if tlsConfig.MinVersion != tls.VersionTLS12 {
		t.Errorf("MinVersion = %v, want %v", tlsConfig.MinVersion, tls.VersionTLS12)
	}

	if tlsConfig.MaxVersion != tls.VersionTLS13 {
		t.Errorf("MaxVersion = %v, want %v", tlsConfig.MaxVersion, tls.VersionTLS13)
	}

	if tlsConfig.GetCertificate == nil {
		t.Error("GetCertificate function is nil")
	}
}

func TestSetCertification_AutoGenerate(t *testing.T) {
	t.Parallel() // 병렬 실행

	tmpDir := t.TempDir()

	config := common.ServerConfig{
		Port: 8443,
		Cert: common.ServerCertConfig{
			AutoGenerate: true,
			Dir:          tmpDir,
			AutoGenerateInfo: common.CertGenerateInfo{
				CommonName:                   "test-server",
				Organization:                 []string{"Test Org"},
				SubjectAlternateNameDnsNames: []string{"localhost"},
				SubjectAlternateNameIps:      []string{"127.0.0.1"},
				NotAfter:                     time.Now().Add(24 * time.Hour).Format(time.RFC3339),
			},
		},
	}

	tlsConfig, err := SetCertification(config)
	if err != nil {
		t.Fatalf("SetCertification() error = %v", err)
	}

	if tlsConfig == nil {
		t.Fatal("SetCertification() returned nil config")
	}

	// Check that certificate files were created
	certFile := filepath.Join(tmpDir, "tls.crt")
	keyFile := filepath.Join(tmpDir, "tls.key")
	caFile := filepath.Join(tmpDir, "tls_ca.crt")

	if _, err := os.Stat(certFile); os.IsNotExist(err) {
		t.Errorf("Certificate file was not created: %s", certFile)
	}
	if _, err := os.Stat(keyFile); os.IsNotExist(err) {
		t.Errorf("Key file was not created: %s", keyFile)
	}
	if _, err := os.Stat(caFile); os.IsNotExist(err) {
		t.Errorf("CA file was not created: %s", caFile)
	}
}

func TestSetCertification_WithDirectory(t *testing.T) {
	t.Parallel() // 병렬 실행

	tmpDir := t.TempDir()

	// Generate and write certificates (빠른 버전 사용)
	notAfter := time.Now().Add(24 * time.Hour)
	caCertPEM, certPEM, keyPEM, err := generateTestPem(
		"test-common-name",
		[]string{"test-org"},
		[]string{"localhost"},
		[]string{"127.0.0.1"},
		notAfter,
	)
	if err != nil {
		t.Fatalf("Failed to generate test certificates: %v", err)
	}

	if err := os.WriteFile(filepath.Join(tmpDir, "tls_ca.crt"), caCertPEM.Bytes(), 0o600); err != nil {
		t.Fatalf("Failed to write CA cert: %v", err)
	}
	if err := os.WriteFile(filepath.Join(tmpDir, "tls.crt"), certPEM.Bytes(), 0o600); err != nil {
		t.Fatalf("Failed to write cert: %v", err)
	}
	if err := os.WriteFile(filepath.Join(tmpDir, "tls.key"), keyPEM.Bytes(), 0o600); err != nil {
		t.Fatalf("Failed to write key: %v", err)
	}

	config := common.ServerConfig{
		Port: 8443,
		Cert: common.ServerCertConfig{
			Dir: tmpDir,
		},
	}

	tlsConfig, err := SetCertification(config)
	if err != nil {
		t.Fatalf("SetCertification() error = %v", err)
	}

	if tlsConfig == nil {
		t.Fatal("SetCertification() returned nil config")
	}
}

func TestSetCertification_NoCertificates(t *testing.T) {
	t.Parallel() // 병렬 실행

	config := common.ServerConfig{
		Port: 8443,
	}

	_, err := SetCertification(config)
	if err == nil {
		t.Error("SetCertification() expected error for missing certificates, got nil")
	}
}

func TestGetCertificationFilePath(t *testing.T) {
	t.Parallel() // 병렬 실행

	tests := []struct {
		name     string
		config   common.ServerConfig
		wantCA   string
		wantCert string
		wantKey  string
	}{
		{
			name: "explicit file paths",
			config: common.ServerConfig{
				Cert: common.ServerCertConfig{
					CaFile:   "/path/to/ca.crt",
					CertFile: "/path/to/cert.crt",
					KeyFile:  "/path/to/key.key",
				},
			},
			wantCA:   "/path/to/ca.crt",
			wantCert: "/path/to/cert.crt",
			wantKey:  "/path/to/key.key",
		},
		{
			name: "directory path",
			config: common.ServerConfig{
				Cert: common.ServerCertConfig{
					Dir: "/path/to/certs",
				},
			},
			wantCA:   filepath.Clean("/path/to/certs") + string(filepath.Separator) + "tls_ca.crt",
			wantCert: filepath.Clean("/path/to/certs") + string(filepath.Separator) + "tls.crt",
			wantKey:  filepath.Clean("/path/to/certs") + string(filepath.Separator) + "tls.key",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel() // 서브테스트도 병렬 실행
			gotCA, gotCert, gotKey := GetCertificationFilePath(tt.config)
			if gotCA != tt.wantCA {
				t.Errorf("GetCertificationFilePath() CA = %v, want %v", gotCA, tt.wantCA)
			}
			if gotCert != tt.wantCert {
				t.Errorf("GetCertificationFilePath() Cert = %v, want %v", gotCert, tt.wantCert)
			}
			if gotKey != tt.wantKey {
				t.Errorf("GetCertificationFilePath() Key = %v, want %v", gotKey, tt.wantKey)
			}
		})
	}
}
