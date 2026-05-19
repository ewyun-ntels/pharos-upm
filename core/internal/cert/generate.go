package cert

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"log/slog"
	"math/big"
	"net"
	"os"
	"path"
	"time"
)

type Config struct {
	CommonName                   string
	Organization                 []string
	SubjectAlternateNameDnsNames []string
	SubjectAlternateNameIps      []string
	NotAfter                     string
	Destination                  string
	KeyFile                      string
	CaCrtFile                    string
	CertFile                     string
}

func GenerateJwtCertAuto(config Config) error {
	jwtKeyPath := path.Join(config.Destination, config.KeyFile)
	if _, err := os.Stat(jwtKeyPath); err == nil {
		return nil
	}

	notAfter, err := time.Parse(time.RFC3339, config.NotAfter)
	if err != nil {
		slog.Error("jwt cert auto generate info not_after parse failed", "error", err)
		return err
	}

	caCertPEM, certPEM, keyPEM, err := GeneratePem(
		config.CommonName,
		config.Organization,
		config.SubjectAlternateNameDnsNames,
		config.SubjectAlternateNameIps,
		notAfter)
	if err != nil {
		slog.Error("jwt cert auto generate failed", "error", err)
		return err
	}

	// Create destination directory with restrictive permissions
	if err := os.MkdirAll(config.Destination, 0o700); err != nil {
		return err
	}

	// Write CA cert and cert with 0600, private key with 0600 to satisfy G306
	if err = os.WriteFile(path.Join(config.Destination, config.CaCrtFile), caCertPEM.Bytes(), 0o600); err != nil {
		return err
	} else if err = os.WriteFile(path.Join(config.Destination, config.CertFile), certPEM.Bytes(), 0o600); err != nil {
		return err
	} else if err = os.WriteFile(jwtKeyPath, keyPEM.Bytes(), 0o600); err != nil {
		return err
	}

	return nil
}

func GeneratePem(commonName string, organization, subjectAlternateNameDNSNames, subjectAlternateNameIps []string, notAfter time.Time) (*bytes.Buffer, *bytes.Buffer, *bytes.Buffer, error) {
	dnsNames := subjectAlternateNameDNSNames
	ipAddresses := []net.IP{}
	for _, ip := range subjectAlternateNameIps {
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

	caCertPrivateKey, err := rsa.GenerateKey(rand.Reader, 4096)
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

	newPrivateKey, err := rsa.GenerateKey(rand.Reader, 4096)
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
