package common

import (
	"sync"
	"testing"
	"time"

	"ntels.com/pharos/core/external/orm"
)

func TestSetGlobalConfig(t *testing.T) {
	// Reset global config for test
	globalConfig = nil
	globalConfigOnce = sync.Once{}

	config := Config{
		Serve: ServeConfig{
			ServerSchema: "https",
		},
		Database: orm.DatabaseConfig{
			Driver: orm.DriverSqlite,
		},
	}

	// Test setting global config
	SetGlobalConfig(config)

	// Verify config was set
	if globalConfig == nil {
		t.Fatal("Expected global config to be set, but it was nil")
	}

	// Test setting config again (should update)
	newConfig := Config{
		Serve: ServeConfig{
			ServerSchema: "http",
		},
	}

	SetGlobalConfig(newConfig)
}

func TestGetGlobalConfig(t *testing.T) {
	// Reset global config for test
	globalConfig = nil
	globalConfigOnce = sync.Once{}

	// Test when no config is set
	result := GetGlobalConfig()
	if result != nil {
		t.Errorf("Expected nil when no config is set, got %v", result)
	}

	// Set a config and test retrieval
	config := Config{
		Serve: ServeConfig{},
	}
	SetGlobalConfig(config)

	result = GetGlobalConfig()
	if result == nil {
		t.Fatal("Expected config to be returned, but got nil")
	}
}

func TestCertGenerateInfo(t *testing.T) {
	certInfo := CertGenerateInfo{
		CommonName:                   "test.example.com",
		Organization:                 []string{"Test Org", "Another Org"},
		SubjectAlternateNameDnsNames: []string{"alt1.example.com", "alt2.example.com"},
		SubjectAlternateNameIps:      []string{"192.168.1.1", "10.0.0.1"},
		NotAfter:                     "2025-12-31",
	}

	if certInfo.CommonName != "test.example.com" {
		t.Errorf("Expected CommonName to be 'test.example.com', got '%s'", certInfo.CommonName)
	}

	if len(certInfo.Organization) != 2 {
		t.Errorf("Expected 2 organizations, got %d", len(certInfo.Organization))
	}

	if len(certInfo.SubjectAlternateNameDnsNames) != 2 {
		t.Errorf("Expected 2 DNS names, got %d", len(certInfo.SubjectAlternateNameDnsNames))
	}

	if len(certInfo.SubjectAlternateNameIps) != 2 {
		t.Errorf("Expected 2 IP addresses, got %d", len(certInfo.SubjectAlternateNameIps))
	}

	if certInfo.NotAfter != "2025-12-31" {
		t.Errorf("Expected NotAfter to be '2025-12-31', got '%s'", certInfo.NotAfter)
	}
}

func TestServeConfig(t *testing.T) {
	serveConfig := ServeConfig{
		ServerSchema: "https",
		NatsSchema:   "nats",
	}

	if serveConfig.ServerSchema != "https" {
		t.Errorf("Expected ServerSchema to be 'https', got '%s'", serveConfig.ServerSchema)
	}

	if serveConfig.NatsSchema != "nats" {
		t.Errorf("Expected NatsSchema to be 'nats', got '%s'", serveConfig.NatsSchema)
	}
}

func TestServerConfig(t *testing.T) {
	serverConfig := ServerConfig{
		Port: 8080,
		Cert: ServerCertConfig{
			AutoGenerate: true,
			Dir:          "/certs",
			KeyFile:      "server.key",
			CertFile:     "server.crt",
			CaFile:       "ca.crt",
		},
		TLSMinVersion: "1.2",
		TLSMaxVersion: "1.3",
	}

	if serverConfig.Port != 8080 {
		t.Errorf("Expected Port to be 8080, got %d", serverConfig.Port)
	}

	if !serverConfig.Cert.AutoGenerate {
		t.Error("Expected Cert.AutoGenerate to be true")
	}

	if serverConfig.TLSMinVersion != "1.2" {
		t.Errorf("Expected TLSMinVersion to be '1.2', got '%s'", serverConfig.TLSMinVersion)
	}

	if serverConfig.TLSMaxVersion != "1.3" {
		t.Errorf("Expected TLSMaxVersion to be '1.3', got '%s'", serverConfig.TLSMaxVersion)
	}
}

func TestPasswordRuleConfig(t *testing.T) {
	passwordRule := PasswordRuleConfig{
		MinLength:           8,
		MinUppercase:        1,
		MinLowercase:        1,
		MinDigits:           1,
		MinSpecial:          1,
		PasswordTTL:         "90d",
		PasswordExpiredRule: []string{"rule1", "rule2"},
	}

	if passwordRule.MinLength != 8 {
		t.Errorf("Expected MinLength to be 8, got %d", passwordRule.MinLength)
	}

	if passwordRule.MinUppercase != 1 {
		t.Errorf("Expected MinUppercase to be 1, got %d", passwordRule.MinUppercase)
	}

	if passwordRule.MinLowercase != 1 {
		t.Errorf("Expected MinLowercase to be 1, got %d", passwordRule.MinLowercase)
	}

	if passwordRule.MinDigits != 1 {
		t.Errorf("Expected MinDigits to be 1, got %d", passwordRule.MinDigits)
	}

	if passwordRule.MinSpecial != 1 {
		t.Errorf("Expected MinSpecial to be 1, got %d", passwordRule.MinSpecial)
	}

	if passwordRule.PasswordTTL != "90d" {
		t.Errorf("Expected PasswordTTL to be '90d', got '%s'", passwordRule.PasswordTTL)
	}

	if len(passwordRule.PasswordExpiredRule) != 2 {
		t.Errorf("Expected 2 password expired rules, got %d", len(passwordRule.PasswordExpiredRule))
	}
}

func TestAuthConfig(t *testing.T) {
	authConfig := AuthConfig{
		GlobalSecret:         "test-secret",
		AccessTokenLifespan:  time.Hour * 24,
		RefreshTokenLifespan: time.Hour * 24 * 30,
		JWTIssuer:            "test-issuer",
		AllowedClients:       []string{"client1", "client2"},
		Clients: []struct {
			ID     string `mapstructure:"id"`
			Secret string `mapstructure:"secret"`
		}{
			{ID: "client1", Secret: "secret1"},
			{ID: "client2", Secret: "secret2"},
		},
	}

	if authConfig.GlobalSecret != "test-secret" {
		t.Errorf("Expected GlobalSecret to be 'test-secret', got '%s'", authConfig.GlobalSecret)
	}

	if authConfig.AccessTokenLifespan != time.Hour*24 {
		t.Errorf("Expected AccessTokenLifespan to be 24h, got %v", authConfig.AccessTokenLifespan)
	}

	if authConfig.RefreshTokenLifespan != time.Hour*24*30 {
		t.Errorf("Expected RefreshTokenLifespan to be 720h, got %v", authConfig.RefreshTokenLifespan)
	}

	if authConfig.JWTIssuer != "test-issuer" {
		t.Errorf("Expected JWTIssuer to be 'test-issuer', got '%s'", authConfig.JWTIssuer)
	}

	if len(authConfig.AllowedClients) != 2 {
		t.Errorf("Expected 2 allowed clients, got %d", len(authConfig.AllowedClients))
	}

	if len(authConfig.Clients) != 2 {
		t.Errorf("Expected 2 clients, got %d", len(authConfig.Clients))
	}
}

func TestStatisticsClickhouseConfig_ToDatabaseConfig(t *testing.T) {
	clickhouseConfig := StatisticsClickhouseConfig{
		Host:              "localhost",
		TcpPort:           9000,
		HttpPort:          8123,
		Schema:            "http",
		Username:          "default",
		Password:          "password",
		Database:          "test_db",
		MaxOpenConnection: 10,
		MaxLifetime:       3600,
	}

	dbConfig := clickhouseConfig.ToDatabaseConfig()

	if dbConfig.Driver != orm.DriverClickHouse {
		t.Errorf("Expected driver to be DriverClickHouse, got %v", dbConfig.Driver)
	}

	if dbConfig.ClickHouse.Host != "localhost" {
		t.Errorf("Expected host to be 'localhost', got '%s'", dbConfig.ClickHouse.Host)
	}

	if dbConfig.ClickHouse.Port != 9000 {
		t.Errorf("Expected port to be 9000, got %d", dbConfig.ClickHouse.Port)
	}

	if dbConfig.ClickHouse.Database != "test_db" {
		t.Errorf("Expected database to be 'test_db', got '%s'", dbConfig.ClickHouse.Database)
	}

	if dbConfig.ClickHouse.MaxOpenConnection != 10 {
		t.Errorf("Expected MaxOpenConnection to be 10, got %d", dbConfig.ClickHouse.MaxOpenConnection)
	}
}

func TestAgentCollectConfig(t *testing.T) {
	collectConfig := AgentCollectConfig{
		Clickhouse: ClickhouseConfig{
			Schema:   "http",
			Host:     "localhost",
			Username: "user",
			Password: "pass",
			Database: "db",
		},
		ServiceDatabase: orm.DatabaseConfig{
			Driver: orm.DriverSqlite,
		},
		Disk: DiskConfig{
			Path: []string{"/path1", "/path2"},
		},
	}

	if collectConfig.Clickhouse.Host != "localhost" {
		t.Errorf("Expected Clickhouse host to be 'localhost', got '%s'", collectConfig.Clickhouse.Host)
	}

	if len(collectConfig.Disk.Path) != 2 {
		t.Errorf("Expected 2 disk paths, got %d", len(collectConfig.Disk.Path))
	}

	if collectConfig.ServiceDatabase.Driver != orm.DriverSqlite {
		t.Errorf("Expected ServiceDatabase driver to be sqlite, got %s", collectConfig.ServiceDatabase.Driver)
	}
}

func TestMetricConfig(t *testing.T) {
	metricConfig := struct {
		Path            string `mapstructure:"path"`
		RetentionPeriod int    `mapstructure:"retention_period"`
	}{
		Path:            "/metrics",
		RetentionPeriod: 7,
	}

	if metricConfig.Path != "/metrics" {
		t.Errorf("Expected Path to be '/metrics', got '%s'", metricConfig.Path)
	}

	if metricConfig.RetentionPeriod != 7 {
		t.Errorf("Expected RetentionPeriod to be 7, got %d", metricConfig.RetentionPeriod)
	}
}

func TestReportConfig(t *testing.T) {
	reportConfig := ReportConfig{
		Use:      true,
		Schedule: "0 0 * * *",
		TTL:      "7d",
		Metric: struct {
			Path            string `mapstructure:"path"`
			RetentionPeriod int    `mapstructure:"retention_period"`
		}{
			Path:            "/data/metrics",
			RetentionPeriod: 3,
		},
	}

	if !reportConfig.Use {
		t.Error("Expected Use to be true")
	}

	if reportConfig.Schedule != "0 0 * * *" {
		t.Errorf("Expected Schedule to be '0 0 * * *', got '%s'", reportConfig.Schedule)
	}

	if reportConfig.TTL != "7d" {
		t.Errorf("Expected TTL to be '7d', got '%s'", reportConfig.TTL)
	}

	if reportConfig.Metric.Path != "/data/metrics" {
		t.Errorf("Expected Metric.Path to be '/data/metrics', got '%s'", reportConfig.Metric.Path)
	}

	if reportConfig.Metric.RetentionPeriod != 3 {
		t.Errorf("Expected Metric.RetentionPeriod to be 3, got %d", reportConfig.Metric.RetentionPeriod)
	}
}

func TestRotationConfig(t *testing.T) {
	rotationConfig := RotationConfig{
		IsJson:           true,
		IsConsole:        false,
		LogPath:          "/var/log/app.log",
		MaxSize:          100,
		MaxBackups:       5,
		MaxAge:           30,
		Compress:         true,
		RotationInterval: "24h",
	}

	if !rotationConfig.IsJson {
		t.Error("Expected IsJson to be true")
	}

	if rotationConfig.IsConsole {
		t.Error("Expected IsConsole to be false")
	}

	if rotationConfig.LogPath != "/var/log/app.log" {
		t.Errorf("Expected LogPath to be '/var/log/app.log', got '%s'", rotationConfig.LogPath)
	}

	if rotationConfig.MaxSize != 100 {
		t.Errorf("Expected MaxSize to be 100, got %d", rotationConfig.MaxSize)
	}

	if rotationConfig.MaxBackups != 5 {
		t.Errorf("Expected MaxBackups to be 5, got %d", rotationConfig.MaxBackups)
	}

	if rotationConfig.MaxAge != 30 {
		t.Errorf("Expected MaxAge to be 30, got %d", rotationConfig.MaxAge)
	}

	if !rotationConfig.Compress {
		t.Error("Expected Compress to be true")
	}

	if rotationConfig.RotationInterval != "24h" {
		t.Errorf("Expected RotationInterval to be '24h', got '%s'", rotationConfig.RotationInterval)
	}
}

func TestMigrationConfig(t *testing.T) {
	migrationConfig := MigrationConfig{
		Disable:      false,
		MigrationDir: "/migrations",
		ArtifactDir:  "/artifacts",
		Env: []MigrationEnv{
			{Name: "ENV1", Value: "value1"},
			{Name: "ENV2", ValueFromEnv: "HOME"},
		},
		Auth: MigrationAuth{
			Id:       "admin",
			Password: "password",
		},
		Wait: "30s",
	}

	if migrationConfig.Disable {
		t.Error("Expected Disable to be false")
	}

	if migrationConfig.MigrationDir != "/migrations" {
		t.Errorf("Expected MigrationDir to be '/migrations', got '%s'", migrationConfig.MigrationDir)
	}

	if migrationConfig.ArtifactDir != "/artifacts" {
		t.Errorf("Expected ArtifactDir to be '/artifacts', got '%s'", migrationConfig.ArtifactDir)
	}

	if len(migrationConfig.Env) != 2 {
		t.Errorf("Expected 2 environment variables, got %d", len(migrationConfig.Env))
	}

	if migrationConfig.Auth.Id != "admin" {
		t.Errorf("Expected auth ID to be 'admin', got '%s'", migrationConfig.Auth.Id)
	}

	if migrationConfig.Wait != "30s" {
		t.Errorf("Expected Wait to be '30s', got '%s'", migrationConfig.Wait)
	}
}

func TestConfig(t *testing.T) {
	config := Config{
		Serve: ServeConfig{},
		Servers: map[string]ServerConfig{
			"main": {
				Port: 8080,
			},
			"admin": {
				Port: 9090,
			},
		},
		Database: orm.DatabaseConfig{
			Driver: orm.DriverSqlite,
		},
		Workflow: WorkflowConfig{
			Use: true,
			Execution: ExecutionConfig{
				TTL: "1h",
			},
		},
		Auth: AuthConfig{
			GlobalSecret: "secret",
		},
		Logger: RotationConfig{
			IsJson: true,
		},
		History: HistoryConfig{
			Use:         true,
			IgnorePaths: []string{"/health", "/metrics"},
		},
		Plugins: PluginsConfig{
			Sample: true,
			Path:   "/plugins",
		},
	}

	if len(config.Servers) != 2 {
		t.Errorf("Expected 2 servers, got %d", len(config.Servers))
	}

	mainServer, exists := config.Servers["main"]
	if !exists {
		t.Error("Expected 'main' server to exist")
	} else if mainServer.Port != 8080 {
		t.Errorf("Expected main server port to be 8080, got %d", mainServer.Port)
	}

	if !config.Workflow.Use {
		t.Error("Expected workflow use to be true")
	}

	if len(config.History.IgnorePaths) != 2 {
		t.Errorf("Expected 2 ignore paths, got %d", len(config.History.IgnorePaths))
	}
}

// TestUserRoleConfig is disabled because UserRoleConfig types were removed during refactoring
/*
func TestUserRoleConfig(t *testing.T) {
	userRoleConfig := UserRoleConfig{
		Composite: UserRoleCompositeConfig{
			"admin": []string{"read", "write", "delete"},
			"user":  []string{"read"},
		},
		DisplayGroup: map[string]UserRoleDisplayGroupConfig{
			"admin": {
				DisplayName: "Administrators",
				Order:       1,
			},
			"user": {
				DisplayName: "Users",
				Order:       2,
			},
		},
		Display: map[string]UserRoleDisplayConfig{
			"admin": {
				DisplayName: "Admin Role",
				Group:       "admin",
			},
		},
	}

	if len(userRoleConfig.Composite) != 2 {
		t.Errorf("Expected 2 composite roles, got %d", len(userRoleConfig.Composite))
	}

	if len(userRoleConfig.DisplayGroup) != 2 {
		t.Errorf("Expected 2 display groups, got %d", len(userRoleConfig.DisplayGroup))
	}

	if len(userRoleConfig.Display) != 1 {
		t.Errorf("Expected 1 display config, got %d", len(userRoleConfig.Display))
	}

	adminGroup, exists := userRoleConfig.DisplayGroup["admin"]
	if !exists {
		t.Error("Expected 'admin' display group to exist")
	} else {
		if adminGroup.DisplayName != "Administrators" {
			t.Errorf("Expected admin display name to be 'Administrators', got '%s'", adminGroup.DisplayName)
		}
		if adminGroup.Order != 1 {
			t.Errorf("Expected admin order to be 1, got %d", adminGroup.Order)
		}
	}

	adminDisplay, exists := userRoleConfig.Display["admin"]
	if !exists {
		t.Error("Expected 'admin' display config to exist")
	} else {
		if adminDisplay.DisplayName != "Admin Role" {
			t.Errorf("Expected admin display name to be 'Admin Role', got '%s'", adminDisplay.DisplayName)
		}
		if adminDisplay.Group != "admin" {
			t.Errorf("Expected admin group to be 'admin', got '%s'", adminDisplay.Group)
		}
	}
}
*/

func TestClientTLSConfig(t *testing.T) {
	clientTLSConfig := ClientTLSConfig{
		InsecureSkipVerify: true,
		MinVersion:         "1.2",
		MaxVersion:         "1.3",
	}

	if !clientTLSConfig.InsecureSkipVerify {
		t.Error("Expected InsecureSkipVerify to be true")
	}

	if clientTLSConfig.MinVersion != "1.2" {
		t.Errorf("Expected MinVersion to be '1.2', got '%s'", clientTLSConfig.MinVersion)
	}

	if clientTLSConfig.MaxVersion != "1.3" {
		t.Errorf("Expected MaxVersion to be '1.3', got '%s'", clientTLSConfig.MaxVersion)
	}
}

func TestProvisioningConfig(t *testing.T) {
	provisioningConfig := ProvisioningConfig{
		Plugins: struct {
			Datasources []map[string]any `mapstructure:"datasources"`
		}{
			Datasources: []map[string]any{
				{
					"name": "prometheus",
					"type": "prometheus",
					"url":  "http://localhost:9090",
				},
				{
					"name": "grafana",
					"type": "grafana",
					"url":  "http://localhost:3000",
				},
			},
		},
	}

	if len(provisioningConfig.Plugins.Datasources) != 2 {
		t.Errorf("Expected 2 datasources, got %d", len(provisioningConfig.Plugins.Datasources))
	}

	firstDatasource := provisioningConfig.Plugins.Datasources[0]
	if firstDatasource["name"] != "prometheus" {
		t.Errorf("Expected first datasource name to be 'prometheus', got '%v'", firstDatasource["name"])
	}
}

// Benchmark tests
func BenchmarkSetGlobalConfig(b *testing.B) {
	config := Config{
		Serve: ServeConfig{},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		SetGlobalConfig(config)
	}
}

func BenchmarkGetGlobalConfig(b *testing.B) {
	config := Config{
		Serve: ServeConfig{},
	}
	SetGlobalConfig(config)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = GetGlobalConfig()
	}
}

func BenchmarkStatisticsClickhouseConfig_ToDatabaseConfig(b *testing.B) {
	clickhouseConfig := StatisticsClickhouseConfig{
		Host:              "localhost",
		TcpPort:           9000,
		HttpPort:          8123,
		Schema:            "http",
		Username:          "default",
		Password:          "password",
		Database:          "test_db",
		MaxOpenConnection: 10,
		MaxLifetime:       3600,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = clickhouseConfig.ToDatabaseConfig()
	}
}

func TestWorkflowConfig(t *testing.T) {
	workflowConfig := WorkflowConfig{
		Use: true,
		Execution: ExecutionConfig{
			TTL: "1h",
		},
	}

	if !workflowConfig.Use {
		t.Error("Expected Use to be true")
	}

	if workflowConfig.Execution.TTL != "1h" {
		t.Errorf("Expected Execution.TTL to be '1h', got '%s'", workflowConfig.Execution.TTL)
	}
}

func TestHistoryConfig(t *testing.T) {
	historyConfig := HistoryConfig{
		Use:         true,
		IgnorePaths: []string{"/health", "/metrics", "/api/v1"},
	}

	if !historyConfig.Use {
		t.Error("Expected Use to be true")
	}

	if len(historyConfig.IgnorePaths) != 3 {
		t.Errorf("Expected 3 ignore paths, got %d", len(historyConfig.IgnorePaths))
	}

	if historyConfig.IgnorePaths[0] != "/health" {
		t.Errorf("Expected first ignore path to be '/health', got '%s'", historyConfig.IgnorePaths[0])
	}
}

func TestPluginsConfig(t *testing.T) {
	pluginsConfig := PluginsConfig{
		Sample: true,
		Path:   "/plugins",
	}

	if !pluginsConfig.Sample {
		t.Error("Expected Sample to be true")
	}

	if pluginsConfig.Path != "/plugins" {
		t.Errorf("Expected Path to be '/plugins', got '%s'", pluginsConfig.Path)
	}
}

func TestClickhouseConfig(t *testing.T) {
	clickhouseConfig := ClickhouseConfig{
		Schema:   "https",
		Host:     "clickhouse.example.com",
		Username: "admin",
		Password: "secure_password",
		Database: "analytics",
	}

	if clickhouseConfig.Schema != "https" {
		t.Errorf("Expected Schema to be 'https', got '%s'", clickhouseConfig.Schema)
	}

	if clickhouseConfig.Host != "clickhouse.example.com" {
		t.Errorf("Expected Host to be 'clickhouse.example.com', got '%s'", clickhouseConfig.Host)
	}

	if clickhouseConfig.Username != "admin" {
		t.Errorf("Expected Username to be 'admin', got '%s'", clickhouseConfig.Username)
	}

	if clickhouseConfig.Password != "secure_password" {
		t.Errorf("Expected Password to be 'secure_password', got '%s'", clickhouseConfig.Password)
	}

	if clickhouseConfig.Database != "analytics" {
		t.Errorf("Expected Database to be 'analytics', got '%s'", clickhouseConfig.Database)
	}
}

func TestDiskConfig(t *testing.T) {
	diskConfig := DiskConfig{
		Path: []string{"/mnt/data1", "/mnt/data2", "/mnt/data3"},
	}

	if len(diskConfig.Path) != 3 {
		t.Errorf("Expected 3 disk paths, got %d", len(diskConfig.Path))
	}

	if diskConfig.Path[1] != "/mnt/data2" {
		t.Errorf("Expected second disk path to be '/mnt/data2', got '%s'", diskConfig.Path[1])
	}
}

func TestServerCertConfig(t *testing.T) {
	certConfig := ServerCertConfig{
		AutoGenerate: true,
		Dir:          "/etc/certs",
		KeyFile:      "server.key",
		CertFile:     "server.crt",
		CaFile:       "ca.crt",
		AutoGenerateInfo: CertGenerateInfo{
			CommonName:   "server.example.com",
			Organization: []string{"Example Corp"},
			NotAfter:     "2030-12-31",
		},
	}

	if !certConfig.AutoGenerate {
		t.Error("Expected AutoGenerate to be true")
	}

	if certConfig.Dir != "/etc/certs" {
		t.Errorf("Expected Dir to be '/etc/certs', got '%s'", certConfig.Dir)
	}

	if certConfig.KeyFile != "server.key" {
		t.Errorf("Expected KeyFile to be 'server.key', got '%s'", certConfig.KeyFile)
	}

	if certConfig.CertFile != "server.crt" {
		t.Errorf("Expected CertFile to be 'server.crt', got '%s'", certConfig.CertFile)
	}

	if certConfig.CaFile != "ca.crt" {
		t.Errorf("Expected CaFile to be 'ca.crt', got '%s'", certConfig.CaFile)
	}

	if certConfig.AutoGenerateInfo.CommonName != "server.example.com" {
		t.Errorf("Expected CommonName to be 'server.example.com', got '%s'", certConfig.AutoGenerateInfo.CommonName)
	}
}

func TestMigrationEnv(t *testing.T) {
	envDirect := MigrationEnv{
		Name:  "DATABASE_URL",
		Value: "postgres://localhost:5432/db",
	}

	envFromEnv := MigrationEnv{
		Name:         "HOME_DIR",
		ValueFromEnv: "HOME",
	}

	if envDirect.Name != "DATABASE_URL" {
		t.Errorf("Expected Name to be 'DATABASE_URL', got '%s'", envDirect.Name)
	}

	if envDirect.Value != "postgres://localhost:5432/db" {
		t.Errorf("Expected Value to be 'postgres://localhost:5432/db', got '%s'", envDirect.Value)
	}

	if envFromEnv.Name != "HOME_DIR" {
		t.Errorf("Expected Name to be 'HOME_DIR', got '%s'", envFromEnv.Name)
	}

	if envFromEnv.ValueFromEnv != "HOME" {
		t.Errorf("Expected ValueFromEnv to be 'HOME', got '%s'", envFromEnv.ValueFromEnv)
	}
}

func TestMigrationAuth(t *testing.T) {
	auth := MigrationAuth{
		Id:       "migration_user",
		Password: "migration_pass",
	}

	if auth.Id != "migration_user" {
		t.Errorf("Expected Id to be 'migration_user', got '%s'", auth.Id)
	}

	if auth.Password != "migration_pass" {
		t.Errorf("Expected Password to be 'migration_pass', got '%s'", auth.Password)
	}
}

func TestExecutionConfig(t *testing.T) {
	execConfig := ExecutionConfig{
		TTL: "2h30m",
	}

	if execConfig.TTL != "2h30m" {
		t.Errorf("Expected TTL to be '2h30m', got '%s'", execConfig.TTL)
	}
}

func TestStatisticsConfig(t *testing.T) {
	statsConfig := StatisticsConfig{
		Database: orm.DatabaseConfig{
			Driver: orm.DriverClickHouse,
			ClickHouse: orm.ClickHouseConfig{
				Host:     "stats.example.com",
				Username: "stats_user",
				Password: "stats_pass",
				Database: "statistics",
			},
		},
	}

	if statsConfig.Database.Driver != orm.DriverClickHouse {
		t.Errorf("Expected Driver to be DriverClickHouse, got %v", statsConfig.Database.Driver)
	}

	if statsConfig.Database.ClickHouse.Host != "stats.example.com" {
		t.Errorf("Expected Host to be 'stats.example.com', got '%s'", statsConfig.Database.ClickHouse.Host)
	}

	if statsConfig.Database.ClickHouse.Username != "stats_user" {
		t.Errorf("Expected Username to be 'stats_user', got '%s'", statsConfig.Database.ClickHouse.Username)
	}

	if statsConfig.Database.ClickHouse.Password != "stats_pass" {
		t.Errorf("Expected Password to be 'stats_pass', got '%s'", statsConfig.Database.ClickHouse.Password)
	}
}

// Test concurrent access to global config
func TestGlobalConfig_Concurrent(t *testing.T) {
	// Reset global config
	globalConfig = nil
	globalConfigOnce = sync.Once{}

	config := Config{
		Serve: ServeConfig{},
	}

	// Set the config once before concurrent reads
	SetGlobalConfig(config)

	// Launch multiple goroutines to read config simultaneously
	var wg sync.WaitGroup
	errors := make(chan error, 10)

	for range 10 {
		wg.Go(func() {
			result := GetGlobalConfig()
			if result == nil {
				errors <- &testError{msg: "Expected config to be set"}
				return
			}
		})
	}

	wg.Wait()
	close(errors)

	for err := range errors {
		if err != nil {
			t.Error(err)
		}
	}
}

type testError struct {
	msg string
}

func (e *testError) Error() string {
	return e.msg
}

// Benchmark tests for new functions
func BenchmarkClickhouseConfig_Fields(b *testing.B) {
	config := ClickhouseConfig{
		Schema:   "https",
		Host:     "localhost",
		Username: "user",
		Password: "pass",
		Database: "db",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = config.Schema + config.Host + config.Database + config.Username + config.Password
	}
}
