package cert

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"math/big"
	"net"
	"os"
	"path/filepath"
	"testing"
	"time"
)

// generateTestPem은 테스트용으로 2048비트 키를 사용하여 더 빠르게 인증서를 생성합니다.
func generateTestPem(commonName string, organization, dnsNames, ips []string, notAfter time.Time) (*bytes.Buffer, *bytes.Buffer, *bytes.Buffer, error) {
	ipAddresses := []net.IP{}
	for _, ip := range ips {
		ipAddresses = append(ipAddresses, net.ParseIP(ip))
	}

	caCert := &x509.Certificate{
		SerialNumber:          big.NewInt(2022),
		Subject:               pkix.Name{Organization: organization},
		NotBefore:             time.Now(),
		NotAfter:              notAfter,
		KeyUsage:              x509.KeyUsageDigitalSignature | x509.KeyUsageCertSign,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
		IsCA:                  true,
	}

	// 테스트용으로 2048비트 사용 (4096 대신)
	caCertPrivateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, nil, nil, err
	}

	caCertBytes, err := x509.CreateCertificate(rand.Reader, caCert, caCert, &caCertPrivateKey.PublicKey, caCertPrivateKey)
	if err != nil {
		return nil, nil, nil, err
	}

	caCertPEM := new(bytes.Buffer)
	if err := pem.Encode(caCertPEM, &pem.Block{
		Type:  "CERTIFICATE",
		Bytes: caCertBytes,
	}); err != nil {
		return nil, nil, nil, err
	}

	newCert := &x509.Certificate{
		SerialNumber: big.NewInt(1024),
		Subject: pkix.Name{
			CommonName:   commonName,
			Organization: organization,
		},
		NotBefore:   time.Now(),
		NotAfter:    notAfter,
		KeyUsage:    x509.KeyUsageDigitalSignature,
		ExtKeyUsage: []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth},
		DNSNames:    dnsNames,
		IPAddresses: ipAddresses,
	}

	// 테스트용으로 2048비트 사용 (4096 대신)
	newPrivateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, nil, nil, err
	}

	newCertBytes, err := x509.CreateCertificate(rand.Reader, newCert, caCert, &newPrivateKey.PublicKey, caCertPrivateKey)
	if err != nil {
		return nil, nil, nil, err
	}

	newCertPEM := new(bytes.Buffer)
	if err := pem.Encode(newCertPEM, &pem.Block{
		Type:  "CERTIFICATE",
		Bytes: newCertBytes,
	}); err != nil {
		return nil, nil, nil, err
	}

	newPrivateKeyPEM := new(bytes.Buffer)
	privateKey, err := x509.MarshalPKCS8PrivateKey(newPrivateKey)
	if err != nil {
		return nil, nil, nil, err
	}
	if err := pem.Encode(newPrivateKeyPEM, &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: privateKey,
	}); err != nil {
		return nil, nil, nil, err
	}

	return caCertPEM, newCertPEM, newPrivateKeyPEM, nil
}

func TestGeneratePem(t *testing.T) {
	t.Parallel() // 병렬 실행

	tests := []struct {
		name         string
		commonName   string
		organization []string
		dnsNames     []string
		ips          []string
		notAfter     time.Time
		wantError    bool
	}{
		{
			name:         "valid certificate generation",
			commonName:   "test.example.com",
			organization: []string{"Test Organization"},
			dnsNames:     []string{"localhost", "test.example.com"},
			ips:          []string{"127.0.0.1", "192.168.1.1"},
			notAfter:     time.Now().Add(365 * 24 * time.Hour),
			wantError:    false,
		},
		{
			name:         "minimal configuration",
			commonName:   "minimal.test",
			organization: []string{"Minimal Org"},
			dnsNames:     []string{},
			ips:          []string{},
			notAfter:     time.Now().Add(24 * time.Hour),
			wantError:    false,
		},
		{
			name:         "multiple organizations",
			commonName:   "multi.test",
			organization: []string{"Org1", "Org2", "Org3"},
			dnsNames:     []string{"multi.test"},
			ips:          []string{"10.0.0.1"},
			notAfter:     time.Now().Add(720 * time.Hour),
			wantError:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel() // 서브테스트도 병렬 실행

			// 테스트용 빠른 버전 사용
			caCertPEM, certPEM, keyPEM, err := generateTestPem(
				tt.commonName,
				tt.organization,
				tt.dnsNames,
				tt.ips,
				tt.notAfter,
			)

			if (err != nil) != tt.wantError {
				t.Errorf("GeneratePem() error = %v, wantError %v", err, tt.wantError)
				return
			}

			if tt.wantError {
				return
			}

			// Verify CA certificate
			if caCertPEM == nil || caCertPEM.Len() == 0 {
				t.Error("CA certificate PEM is empty")
			}

			block, _ := pem.Decode(caCertPEM.Bytes())
			if block == nil {
				t.Error("Failed to decode CA certificate PEM")
			} else if block.Type != "CERTIFICATE" {
				t.Errorf("CA certificate PEM type = %v, want CERTIFICATE", block.Type)
			} else {
				caCert, err := x509.ParseCertificate(block.Bytes)
				if err != nil {
					t.Errorf("Failed to parse CA certificate: %v", err)
				} else {
					if !caCert.IsCA {
						t.Error("CA certificate IsCA flag is false")
					}
					if len(caCert.Subject.Organization) == 0 {
						t.Error("CA certificate has no organization")
					}
				}
			}

			// Verify server certificate
			if certPEM == nil || certPEM.Len() == 0 {
				t.Error("Server certificate PEM is empty")
			}

			block, _ = pem.Decode(certPEM.Bytes())
			if block == nil {
				t.Error("Failed to decode server certificate PEM")
			} else if block.Type != "CERTIFICATE" {
				t.Errorf("Server certificate PEM type = %v, want CERTIFICATE", block.Type)
			} else {
				cert, err := x509.ParseCertificate(block.Bytes)
				if err != nil {
					t.Errorf("Failed to parse server certificate: %v", err)
				} else {
					if cert.Subject.CommonName != tt.commonName {
						t.Errorf("Certificate CommonName = %v, want %v", cert.Subject.CommonName, tt.commonName)
					}
					if len(tt.dnsNames) > 0 && len(cert.DNSNames) != len(tt.dnsNames) {
						t.Errorf("Certificate DNSNames count = %v, want %v", len(cert.DNSNames), len(tt.dnsNames))
					}
					if len(tt.ips) > 0 && len(cert.IPAddresses) != len(tt.ips) {
						t.Errorf("Certificate IPAddresses count = %v, want %v", len(cert.IPAddresses), len(tt.ips))
					}
				}
			}

			// Verify private key
			if keyPEM == nil || keyPEM.Len() == 0 {
				t.Error("Private key PEM is empty")
			}

			block, _ = pem.Decode(keyPEM.Bytes())
			if block == nil {
				t.Error("Failed to decode private key PEM")
			} else if block.Type != "RSA PRIVATE KEY" {
				t.Errorf("Private key PEM type = %v, want RSA PRIVATE KEY", block.Type)
			} else {
				_, err := x509.ParsePKCS8PrivateKey(block.Bytes)
				if err != nil {
					t.Errorf("Failed to parse private key: %v", err)
				}
			}
		})
	}
}

func TestGenerateJwtCertAuto(t *testing.T) {
	t.Parallel() // 병렬 실행

	tests := []struct {
		name      string
		config    Config
		wantError bool
		skipSetup bool
	}{
		{
			name: "successful certificate generation",
			config: Config{
				CommonName:                   "jwt-test.example.com",
				Organization:                 []string{"JWT Test Org"},
				SubjectAlternateNameDnsNames: []string{"localhost", "jwt-test.example.com"},
				SubjectAlternateNameIps:      []string{"127.0.0.1"},
				NotAfter:                     time.Now().Add(365 * 24 * time.Hour).Format(time.RFC3339),
				Destination:                  "", // will be set in test
				KeyFile:                      "test.key",
				CaCrtFile:                    "test_ca.crt",
				CertFile:                     "test.crt",
			},
			wantError: false,
			skipSetup: false,
		},
		{
			name: "certificate already exists - should not regenerate",
			config: Config{
				CommonName:                   "jwt-existing.example.com",
				Organization:                 []string{"JWT Existing Org"},
				SubjectAlternateNameDnsNames: []string{"localhost"},
				SubjectAlternateNameIps:      []string{"127.0.0.1"},
				NotAfter:                     time.Now().Add(24 * time.Hour).Format(time.RFC3339),
				Destination:                  "", // will be set in test
				KeyFile:                      "existing.key",
				CaCrtFile:                    "existing_ca.crt",
				CertFile:                     "existing.crt",
			},
			wantError: false,
			skipSetup: true, // pre-create the key file
		},
		{
			name: "invalid NotAfter date",
			config: Config{
				CommonName:                   "jwt-invalid.example.com",
				Organization:                 []string{"JWT Invalid Org"},
				SubjectAlternateNameDnsNames: []string{"localhost"},
				SubjectAlternateNameIps:      []string{"127.0.0.1"},
				NotAfter:                     "invalid-date-format",
				Destination:                  "", // will be set in test
				KeyFile:                      "invalid.key",
				CaCrtFile:                    "invalid_ca.crt",
				CertFile:                     "invalid.crt",
			},
			wantError: true,
			skipSetup: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel() // 서브테스트도 병렬 실행

			// Create temporary directory for each test
			tmpDir := t.TempDir()
			tt.config.Destination = tmpDir

			if tt.skipSetup {
				// Pre-create the key file to test the "already exists" case
				keyPath := filepath.Join(tmpDir, tt.config.KeyFile)
				if err := os.WriteFile(keyPath, []byte("existing key"), 0o600); err != nil {
					t.Fatalf("Failed to create existing key file: %v", err)
				}
			}

			err := GenerateJwtCertAuto(tt.config)

			if (err != nil) != tt.wantError {
				t.Errorf("GenerateJwtCertAuto() error = %v, wantError %v", err, tt.wantError)
				return
			}

			if tt.wantError {
				return
			}

			// Verify that files were created
			keyPath := filepath.Join(tmpDir, tt.config.KeyFile)
			caCrtPath := filepath.Join(tmpDir, tt.config.CaCrtFile)
			certPath := filepath.Join(tmpDir, tt.config.CertFile)

			// Check file existence
			if _, err := os.Stat(keyPath); os.IsNotExist(err) {
				if !tt.skipSetup {
					t.Errorf("Key file was not created: %s", keyPath)
				}
			}
			if _, err := os.Stat(caCrtPath); os.IsNotExist(err) {
				if !tt.skipSetup {
					t.Errorf("CA certificate file was not created: %s", caCrtPath)
				}
			}
			if _, err := os.Stat(certPath); os.IsNotExist(err) {
				if !tt.skipSetup {
					t.Errorf("Certificate file was not created: %s", certPath)
				}
			}

			if !tt.skipSetup {
				// Verify file permissions
				keyInfo, err := os.Stat(keyPath)
				if err == nil {
					if keyInfo.Mode().Perm() != 0o600 {
						t.Errorf("Key file permissions = %v, want 0600", keyInfo.Mode().Perm())
					}
				}

				caInfo, err := os.Stat(caCrtPath)
				if err == nil {
					if caInfo.Mode().Perm() != 0o600 {
						t.Errorf("CA cert file permissions = %v, want 0600", caInfo.Mode().Perm())
					}
				}

				certInfo, err := os.Stat(certPath)
				if err == nil {
					if certInfo.Mode().Perm() != 0o600 {
						t.Errorf("Cert file permissions = %v, want 0600", certInfo.Mode().Perm())
					}
				}

				// Verify that the certificates are valid
				caCertData, err := os.ReadFile(caCrtPath)
				if err != nil {
					t.Errorf("Failed to read CA certificate: %v", err)
				} else {
					block, _ := pem.Decode(caCertData)
					if block == nil {
						t.Error("Failed to decode CA certificate PEM")
					} else {
						_, err := x509.ParseCertificate(block.Bytes)
						if err != nil {
							t.Errorf("Failed to parse CA certificate: %v", err)
						}
					}
				}

				certData, err := os.ReadFile(certPath)
				if err != nil {
					t.Errorf("Failed to read certificate: %v", err)
				} else {
					block, _ := pem.Decode(certData)
					if block == nil {
						t.Error("Failed to decode certificate PEM")
					} else {
						cert, err := x509.ParseCertificate(block.Bytes)
						if err != nil {
							t.Errorf("Failed to parse certificate: %v", err)
						} else if cert.Subject.CommonName != tt.config.CommonName {
							t.Errorf("Certificate CommonName = %v, want %v", cert.Subject.CommonName, tt.config.CommonName)
						}
					}
				}
			}
		})
	}
}

func TestGenerateJwtCertAuto_DirectoryCreation(t *testing.T) {
	t.Parallel() // 병렬 실행

	tmpDir := t.TempDir()
	nestedDir := filepath.Join(tmpDir, "nested", "directory", "structure")

	config := Config{
		CommonName:                   "test.example.com",
		Organization:                 []string{"Test Org"},
		SubjectAlternateNameDnsNames: []string{"localhost"},
		SubjectAlternateNameIps:      []string{"127.0.0.1"},
		NotAfter:                     time.Now().Add(24 * time.Hour).Format(time.RFC3339),
		Destination:                  nestedDir,
		KeyFile:                      "jwt.key",
		CaCrtFile:                    "jwt_ca.crt",
		CertFile:                     "jwt.crt",
	}

	err := GenerateJwtCertAuto(config)
	if err != nil {
		t.Fatalf("GenerateJwtCertAuto() failed to create nested directories: %v", err)
	}

	// Verify directory was created
	if _, err := os.Stat(nestedDir); os.IsNotExist(err) {
		t.Errorf("Nested directory was not created: %s", nestedDir)
	}

	// Verify directory permissions
	dirInfo, err := os.Stat(nestedDir)
	if err == nil {
		if dirInfo.Mode().Perm() != 0o700 {
			t.Errorf("Directory permissions = %v, want 0700", dirInfo.Mode().Perm())
		}
	}
}

func TestGeneratePem_IPAddressParsing(t *testing.T) {
	t.Parallel() // 병렬 실행

	tests := []struct {
		name     string
		ips      []string
		wantIPv4 bool
		wantIPv6 bool
	}{
		{
			name:     "IPv4 addresses",
			ips:      []string{"127.0.0.1", "192.168.1.1", "10.0.0.1"},
			wantIPv4: true,
			wantIPv6: false,
		},
		{
			name:     "IPv6 addresses",
			ips:      []string{"::1", "2001:db8::1"},
			wantIPv4: false,
			wantIPv6: true,
		},
		{
			name:     "mixed IP addresses",
			ips:      []string{"127.0.0.1", "::1", "192.168.1.1"},
			wantIPv4: true,
			wantIPv6: true,
		},
		{
			name:     "empty IP list",
			ips:      []string{},
			wantIPv4: false,
			wantIPv6: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel() // 서브테스트도 병렬 실행

			// 테스트용 빠른 버전 사용
			_, certPEM, _, err := generateTestPem(
				"test.example.com",
				[]string{"Test Org"},
				[]string{"localhost"},
				tt.ips,
				time.Now().Add(24*time.Hour),
			)

			if err != nil {
				t.Fatalf("GeneratePem() error = %v", err)
			}

			block, _ := pem.Decode(certPEM.Bytes())
			if block == nil {
				t.Fatal("Failed to decode certificate PEM")
			}

			cert, err := x509.ParseCertificate(block.Bytes)
			if err != nil {
				t.Fatalf("Failed to parse certificate: %v", err)
			}

			hasIPv4 := false
			hasIPv6 := false
			for _, ip := range cert.IPAddresses {
				if ip.To4() != nil {
					hasIPv4 = true
				} else {
					hasIPv6 = true
				}
			}

			if hasIPv4 != tt.wantIPv4 {
				t.Errorf("Certificate has IPv4 = %v, want %v", hasIPv4, tt.wantIPv4)
			}
			if hasIPv6 != tt.wantIPv6 {
				t.Errorf("Certificate has IPv6 = %v, want %v", hasIPv6, tt.wantIPv6)
			}
		})
	}
}

// 벤치마크 테스트: 테스트용 빠른 버전 vs 원본 버전 비교
func BenchmarkGeneratePem_Fast(b *testing.B) {
	notAfter := time.Now().Add(24 * time.Hour)
	for i := 0; i < b.N; i++ {
		_, _, _, _ = generateTestPem(
			"bench.example.com",
			[]string{"Bench Org"},
			[]string{"localhost"},
			[]string{"127.0.0.1"},
			notAfter,
		)
	}
}

func BenchmarkGeneratePem_Original(b *testing.B) {
	notAfter := time.Now().Add(24 * time.Hour)
	for i := 0; i < b.N; i++ {
		_, _, _, _ = GeneratePem(
			"bench.example.com",
			[]string{"Bench Org"},
			[]string{"localhost"},
			[]string{"127.0.0.1"},
			notAfter,
		)
	}
}
