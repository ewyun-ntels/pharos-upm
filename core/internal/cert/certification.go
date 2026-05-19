package cert

import (
	"crypto/tls"
	"errors"
	"os"
	"path/filepath"
	"strings"

	"ntels.com/pharos/core/pkg/common"
)

func parseTLSVersion(ver string) (uint16, bool) {
	switch ver {
	case "", "default":
		return tls.VersionTLS13, false
	case "1.2", "TLS1.2", "tls1.2":
		return tls.VersionTLS12, true
	case "1.3", "TLS1.3", "tls1.3":
		return tls.VersionTLS13, true
	default:
		return 0, false
	}
}

func resolveTLSVersions(cfg common.ServerConfig) (min uint16, max uint16) {
	// Defaults: 1.3
	min = tls.VersionTLS13
	max = tls.VersionTLS13
	if v, ok := parseTLSVersion(cfg.TLSMinVersion); ok {
		min = v
	}
	if v, ok := parseTLSVersion(cfg.TLSMaxVersion); ok {
		max = v
	}
	return
}

const (
	tlsCaCrt = "tls_ca.crt"
	tlsCrt   = "tls.crt"
	tlsKey   = "tls.key"
)

func safeReadFile(baseDir, p string) ([]byte, error) {
	clean := filepath.Clean(p)
	if baseDir != "" {
		absBase, _ := filepath.Abs(baseDir)
		absFile, _ := filepath.Abs(clean)
		if !strings.HasPrefix(absFile, absBase+string(filepath.Separator)) && absFile != absBase {
			return nil, errors.New("file path outside of allowed directory")
		}
	}
	return os.ReadFile(clean)
}

func getCertificationFunc(certFile, keyFile string) func(*tls.ClientHelloInfo) (*tls.Certificate, error) {
	baseDir := filepath.Dir(certFile)
	return func(*tls.ClientHelloInfo) (*tls.Certificate, error) {
		if readCert, err := safeReadFile(baseDir, certFile); err != nil {
			return nil, err
		} else if readKey, err := safeReadFile(baseDir, keyFile); err != nil {
			return nil, err
		} else {
			certificate, err := tls.X509KeyPair(readCert, readKey)
			return &certificate, err
		}
	}
}

func SetCertification(server common.ServerConfig) (*tls.Config, error) {
	minVer, maxVer := resolveTLSVersions(server)

	if server.Cert.KeyFile != "" && server.Cert.CertFile != "" {
		// #nosec G402 TLS 하위호환성 고려
		return &tls.Config{
			GetCertificate: getCertificationFunc(server.Cert.CertFile, server.Cert.KeyFile),
			MinVersion:     minVer,
			MaxVersion:     maxVer,
		}, nil
	} else if server.Cert.Dir != "" && server.Cert.AutoGenerate {
		keyPath := filepath.Clean(server.Cert.Dir) + string(filepath.Separator) + tlsKey
		certPath := filepath.Clean(server.Cert.Dir) + string(filepath.Separator) + tlsCrt

		if _, err := os.Stat(keyPath); os.IsExist(err) {
			// #nosec G402 TLS 하위호환성 고려
			return &tls.Config{
				GetCertificate: getCertificationFunc(certPath, keyPath),
				MinVersion:     minVer,
				MaxVersion:     maxVer,
			}, nil
		}

		err := GenerateJwtCertAuto(Config{
			CommonName:                   server.Cert.AutoGenerateInfo.CommonName,
			Organization:                 server.Cert.AutoGenerateInfo.Organization,
			SubjectAlternateNameDnsNames: server.Cert.AutoGenerateInfo.SubjectAlternateNameDnsNames,
			SubjectAlternateNameIps:      server.Cert.AutoGenerateInfo.SubjectAlternateNameIps,
			NotAfter:                     server.Cert.AutoGenerateInfo.NotAfter,
			Destination:                  server.Cert.Dir,
			KeyFile:                      tlsKey,
			CaCrtFile:                    tlsCaCrt,
			CertFile:                     tlsCrt,
		})
		if err != nil {
			return nil, err
		}

		// #nosec G402 TLS 하위호환성 고려
		return &tls.Config{
			GetCertificate: getCertificationFunc(
				certPath,
				keyPath,
			),
			MinVersion: minVer,
			MaxVersion: maxVer,
		}, nil
	} else if server.Cert.Dir != "" {
		keyPath := filepath.Clean(server.Cert.Dir) + string(filepath.Separator) + tlsKey
		certPath := filepath.Clean(server.Cert.Dir) + string(filepath.Separator) + tlsCrt

		// #nosec G402 TLS 하위호환성 고려
		return &tls.Config{
			GetCertificate: getCertificationFunc(
				certPath,
				keyPath,
			),
			MinVersion: minVer,
			MaxVersion: maxVer,
		}, nil
	}

	return nil, errors.New("cert not found")
}

func GetCertificationFilePath(server common.ServerConfig) (ca, cert, key string) {
	if server.Cert.CaFile != "" && server.Cert.KeyFile != "" && server.Cert.CertFile != "" {
		ca = server.Cert.CaFile
		cert = server.Cert.CertFile
		key = server.Cert.KeyFile
	} else {
		ca = filepath.Clean(server.Cert.Dir) + string(filepath.Separator) + tlsCaCrt
		cert = filepath.Clean(server.Cert.Dir) + string(filepath.Separator) + tlsCrt
		key = filepath.Clean(server.Cert.Dir) + string(filepath.Separator) + tlsKey
	}

	return
}
