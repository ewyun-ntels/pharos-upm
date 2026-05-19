package common

import (
	"crypto/tls"
	"fmt"
	"slices"
	"strings"

	"github.com/go-resty/resty/v2"
)

// parseTLSVersion converts a string like "1.2" or "1.3" to the tls.Version* constant.
// Returns (0, false) if unknown.
func parseTLSVersion(ver string) (uint16, bool) {
	switch strings.TrimSpace(strings.ToLower(ver)) {
	case "", "default":
		return tls.VersionTLS13, false
	case "1.2", "tls1.2":
		return tls.VersionTLS12, true
	case "1.3", "tls1.3":
		return tls.VersionTLS13, true
	default:
		return 0, false
	}
}

// BuildDefaultClientTLS builds a TLS config for outbound clients with secure defaults.
// Priority of settings: Config (if loaded) > Defaults
// Note: Environment variables are intentionally ignored here by design.
func BuildDefaultClientTLS() *tls.Config {
	// Defaults support TLS 1.2..1.3
	insecure := false
	var minVersion uint16 = tls.VersionTLS12
	var maxVersion uint16 = tls.VersionTLS13

	// Load from global config if available
	if cfg := GetGlobalConfig(); cfg != nil {
		// insecure skip verify
		insecure = cfg.ClientTLS.InsecureSkipVerify
		// min/max version from config (strings like "1.2", "1.3")
		if v, ok := parseTLSVersion(cfg.ClientTLS.MinVersion); ok {
			minVersion = v
		}
		if v, ok := parseTLSVersion(cfg.ClientTLS.MaxVersion); ok {
			maxVersion = v
		}
	}

	// Ensure min<=max to avoid invalid config
	if maxVersion != 0 && minVersion != 0 && maxVersion < minVersion {
		maxVersion = minVersion
	}

	// #nosec G402 TLS 하위호환성 고려 및 사설 인증서 연동 필요
	return &tls.Config{InsecureSkipVerify: insecure, MinVersion: minVersion, MaxVersion: maxVersion}
}

func GetRestyClient() *resty.Client {
	client := resty.New()
	client.SetTLSClientConfig(BuildDefaultClientTLS())
	return client
}

func GetClickhouseMasterURL(host string) string {
	return host + "/proxy/clickhouse"
}

func GetWebsocketEndpoints(config Config) []string {
	var endpointsHttp []string
	var endpointsHttps []string

	for schema, server := range config.Servers {
		if schema != config.Serve.ServerSchema {
			continue
		}

		endpoint := fmt.Sprintf("://localhost:%d/websocket/centrifuge", server.Port)

		switch schema {
		case SchemaHttp:
			endpointsHttp = append(endpointsHttp, "ws"+endpoint)
		case SchemaHttps:
			endpointsHttps = append(endpointsHttps, "wss"+endpoint)
		}
	}

	return slices.Concat(endpointsHttp, endpointsHttps)
}
