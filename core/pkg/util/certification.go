package util

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/cobra"
	coreV1 "k8s.io/api/core/v1"
	"ntels.com/pharos/core/internal/cert"
)

var certification = &cobra.Command{
	Use:   "certification subject-alternate-name-dns-names subject-alternate-name-ips not-after destination",
	Short: "certification",
	PreRunE: func(cmd *cobra.Command, args []string) error {
		if len(args) != 4 {
			return fmt.Errorf("invalid args(%v)", args)
		} else if _, err := time.Parse(time.DateOnly, args[2]); err != nil {
			return err
		}

		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		subjectAlternateNameDNSNames := strings.Split(args[0], ",")
		subjectAlternateNameIps := strings.Split(args[1], ",")
		notAfter, _ := time.Parse(time.DateOnly, args[2])
		destination := filepath.Clean(args[3]) + string(filepath.Separator)

		if err := os.MkdirAll(destination, 0o700); err != nil {
			return err
		} else if caCertPEM, certPEM, keyPEM, err := cert.GeneratePem("pharos", []string{"tazan.com"},
			subjectAlternateNameDNSNames, subjectAlternateNameIps, notAfter); err != nil {
			return err
		} else if err := os.WriteFile(destination+coreV1.ServiceAccountRootCAKey, caCertPEM.Bytes(), 0o600); err != nil {
			return err
		} else if err := os.WriteFile(destination+coreV1.TLSCertKey, certPEM.Bytes(), 0o600); err != nil {
			return err
		} else if err := os.WriteFile(destination+coreV1.TLSPrivateKeyKey, keyPEM.Bytes(), 0o600); err != nil {
			return err
		}

		return nil
	},
}
