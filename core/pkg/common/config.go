package common

import (
	"sync"
	"time"

	"ntels.com/pharos/core/external/orm"
)

var (
	globalConfig     *Config
	globalConfigOnce sync.Once
)

type AuthHintLevel int

var (
	AuthenticateHintBlock = AuthHintLevel(1)
	AuthenticateHintAll   = AuthHintLevel(2)
)

// TODO Config를 현재 각자 관리하는 방식대신 중앙에서 통합 관리 하는 방식으로 변경 필요

// SetGlobalConfig stores a pointer to the loaded configuration for global access
func SetGlobalConfig(cfg Config) {
	// store a copy to avoid external mutation; we allocate new and copy
	c := cfg
	globalConfigOnce.Do(func() {
		globalConfig = &c
	})
	// If already set, update it atomically (best-effort without sync primitives since reads are racy-safe for pointer swap)
	if globalConfig != &c {
		globalConfig = &c
	}
}

// GetGlobalConfig returns the globally stored configuration if set
func GetGlobalConfig() *Config {
	return globalConfig
}

type CertGenerateInfo struct {
	CommonName                   string   `mapstructure:"common_name"`
	Organization                 []string `mapstructure:"organization"`
	SubjectAlternateNameDnsNames []string `mapstructure:"subject_alternate_name_dns_names"`
	SubjectAlternateNameIps      []string `mapstructure:"subject_alternate_name_ips"`
	NotAfter                     string   `mapstructure:"not_after"`
}

type ServeConfig struct {
	ServerSchema string `mapstructure:"server_schema"`
	NatsSchema   string `mapstructure:"nats_schema"`
}

type ServerCertConfig struct {
	AutoGenerate       bool             `mapstructure:"auto_generate"`
	Dir                string           `mapstructure:"dir"`
	KeyFile            string           `mapstructure:"key_file"`
	CertFile           string           `mapstructure:"cert_file"`
	CaFile             string           `mapstructure:"ca_file"`
	InsecureSkipVerify bool             `mapstructure:"insecure_skip_verify"`
	AutoGenerateInfo   CertGenerateInfo `mapstructure:"auto_generate_info"`
}

type ServerConfig struct {
	Port          int              `mapstructure:"port"`
	Cert          ServerCertConfig `mapstructure:"cert"`
	TLSMinVersion string           `mapstructure:"tls_min_version"`
	TLSMaxVersion string           `mapstructure:"tls_max_version"`
}

type ClientConfig struct {
	URL      string `mapstructure:"url"`
	KeyFile  string `mapstructure:"key_file"`
	CertFile string `mapstructure:"cert_file"`
	CaFile   string `mapstructure:"ca_file"`
}

type PasswordRuleConfig struct {
	MinLength           int      `mapstructure:"min_length"`
	MinUppercase        int      `mapstructure:"min_uppercase"`
	MinLowercase        int      `mapstructure:"min_lowercase"`
	MinDigits           int      `mapstructure:"min_digits"`
	MinSpecial          int      `mapstructure:"min_special"`
	PasswordTTL         string   `mapstructure:"password_ttl"`
	PasswordExpiredRule []string `mapstructure:"password_expired_rule"`
}

type UserRuleConfig struct {
	EmailAllowed bool `mapstructure:"email_allowed"`
	MinLength    int  `mapstructure:"min_length"`
	MaxLength    int  `mapstructure:"max_length"`
}

// CompositeRole - TOML config 로딩을 위한 래퍼
// shared/types/role.Role과 호환되는 구조
// Note: Config file loading removed - this is only for backward compatibility
type CompositeRole struct {
	Roles       map[string]bool `mapstructure:"roles" json:"roles,omitempty"`
	DisplayName string          `mapstructure:"display_name" json:"display_name"`
	Group       *string         `mapstructure:"group" json:"group,omitempty"`
}

type AuthJwtCert struct {
	AutoGenerate     bool             `mapstructure:"auto_generate"`
	JwtDir           string           `mapstructure:"jwt_dir"`
	JwtKeyFile       string           `mapstructure:"jwt_key_file"`
	AutoGenerateInfo CertGenerateInfo `mapstructure:"auto_generate_info"`
}

type AuthConfig struct {
	AuthenticateHint     AuthHintLevel `mapstructure:"authenticate_hint"`
	CookieMode           bool          `mapstructure:"cookie_mode"`
	CsrfProtection       bool          `mapstructure:"csrf_protection"`
	SecureCookies        bool          `mapstructure:"secure_cookies"`
	CollectClientInfo    bool          `mapstructure:"collect_client_info"`
	GlobalSecret         string        `mapstructure:"global_secret"`
	AccessTokenLifespan  time.Duration `mapstructure:"access_token_lifespan"`
	RefreshTokenLifespan time.Duration `mapstructure:"refresh_token_lifespan"`
	JWTIssuer            string        `mapstructure:"jwt_issuer"`
	AllowedClients       []string      `mapstructure:"allowed_clients"`
	JwtCert              AuthJwtCert   `mapstructure:"jwt_cert"`
	Clients              []struct {
		ID     string `mapstructure:"id"`
		Secret string `mapstructure:"secret"`
	} `mapstructure:"clients"`
}

type UsersConfig struct {
	Prepare                  []string           `mapstructure:"prepare"`
	LoginRetryLimit          int                `mapstructure:"login_retry_limit"`
	LoginRetryResetTimeout   int                `mapstructure:"login_retry_reset_timeout"`
	LoginBlockDuration       int                `mapstructure:"login_block_duration"`
	PasswordHistoryRetention int                `mapstructure:"password_history_retention"`
	PasswordRule             PasswordRuleConfig `mapstructure:"password_rule"`
	UserRule                 UserRuleConfig     `mapstructure:"user_rule"`
}

type DiskConfig struct {
	Path []string `mapstructure:"path"`
}

// ClickhouseConfig deprecated
type ClickhouseConfig struct {
	Schema   string `mapstructure:"schema"`
	Host     string `mapstructure:"host"`
	Username string `mapstructure:"username"`
	Password string `mapstructure:"password"`
	Database string `mapstructure:"database"`
}

type AgentCollectConfig struct {
	// TODO master url로 변경필요
	Clickhouse      ClickhouseConfig   `mapstructure:"clickhouse"`
	ClickhouseHttp  ClickhouseConfig   `mapstructure:"clickhouse_http"`
	ServiceDatabase orm.DatabaseConfig `mapstructure:"service_database"`
	Disk            DiskConfig         `mapstructure:"disk"`
}

type ExecutionConfig struct {
	TTL string `mapstructure:"ttl"`
}

type WorkflowConfig struct {
	Use bool `mapstructure:"use"`

	Execution ExecutionConfig `mapstructure:"execution"`
}

type StatisticsClickhouseConfig struct {
	Host              string `mapstructure:"host"`
	TcpPort           int    `mapstructure:"tcp_port"`
	HttpPort          int    `mapstructure:"http_port"`
	Schema            string `mapstructure:"schema"`
	Username          string `mapstructure:"username"`
	Password          string `mapstructure:"password"`
	Database          string `mapstructure:"database"`
	MaxOpenConnection int    `mapstructure:"max_open_connection"`
	MaxLifetime       int    `mapstructure:"max_lifetime"`
}

func (statisticsClickhouseConfig *StatisticsClickhouseConfig) ToDatabaseConfig() orm.DatabaseConfig {
	return orm.DatabaseConfig{
		Driver: orm.DriverClickHouse,
		ClickHouse: orm.ClickHouseConfig{
			Host:              statisticsClickhouseConfig.Host,
			Port:              statisticsClickhouseConfig.TcpPort,
			Username:          statisticsClickhouseConfig.Username,
			Password:          statisticsClickhouseConfig.Password,
			Database:          statisticsClickhouseConfig.Database,
			MaxOpenConnection: statisticsClickhouseConfig.MaxOpenConnection,
			MaxLifetime:       statisticsClickhouseConfig.MaxLifetime,
		},
	}
}

type StatisticsConfig struct {
	// New top-level statistics database config: [statistics.database]
	Database orm.DatabaseConfig `mapstructure:"database"`
}

type MasterConfig struct {
	Statistics StatisticsConfig `mapstructure:"statistics"`
}

type RotationConfig struct {
	Level            string `mapstructure:"level"`             // Log level (e.g., "debug", "info", "warn", "error")
	IsJson           bool   `mapstructure:"json"`              // Use JSON format for logs
	IsConsole        bool   `mapstructure:"console"`           // Log to console
	FileName         string `mapstructure:"file_name"`         // Log file name (e.g., "app.log"), not the full path
	LogPath          string `mapstructure:"log_path"`          // Log file path (e.g. "./logs/app.log")
	MaxSize          int    `mapstructure:"max_size"`          // MB before rotation
	MaxBackups       int    `mapstructure:"max_backups"`       // Max old files to retain
	MaxAge           int    `mapstructure:"max_age"`           // Days to retain old files
	Compress         bool   `mapstructure:"compress"`          // Gzip compress rotated files
	RotationInterval string `mapstructure:"rotation_interval"` // Duration string for forced rotation (e.g. "24h", "2h30m")
}

type HistoryConfig struct {
	Use         bool     `mapstructure:"use"`
	IgnorePaths []string `mapstructure:"ignore_paths"`
}

type PluginsConfig struct {
	Sample bool   `mapstructure:"sample"`
	Path   string `mapstructure:"path"`
}

type CollectStatisticsConfig struct {
	Use      bool   `mapstructure:"use"`
	Schedule string `mapstructure:"schedule"`

	Report ReportConfig `mapstructure:"report"`
}

type CollectConfig struct {
	Statistics CollectStatisticsConfig `mapstructure:"statistics"`
}

type ReportConfig struct {
	Use      bool   `mapstructure:"use"`
	Schedule string `mapstructure:"schedule"`
	TTL      string `mapstructure:"ttl"`

	Metric struct {
		Path            string `mapstructure:"path"`
		RetentionPeriod int    `mapstructure:"retention_period"`
	} `mapstructure:"metric"`
}

type SharedLibraryConfig struct {
	Altibase map[int]struct {
		Driver string `mapstructure:"driver"`
	} `mapstructure:"altibase"`
}

type ProvisioningConfig struct {
	Plugins struct {
		Datasources []map[string]any `mapstructure:"datasources"`
	} `mapstructure:"plugins"`
}

type MigrationEnv struct {
	Name         string `mapstructure:"name"`
	Value        string `mapstructure:"value"`
	ValueFromEnv string `mapstructure:"valueFromEnv"`
}

type MigrationAuth struct {
	Id       string `mapstructure:"id"`
	Password string `mapstructure:"password"`
}

type MigrationConfig struct { // note: key is intentionally spelled as in config
	Disable      bool           `mapstructure:"disable"`
	MigrationDir string         `mapstructure:"migration_dir"`
	ArtifactDir  string         `mapstructure:"artifact_dir"`
	Env          []MigrationEnv `mapstructure:"env"`
	Auth         MigrationAuth  `mapstructure:"auth"`
	Wait         string         `mapstructure:"wait"`
}

type ClientTLSConfig struct {
	InsecureSkipVerify bool   `mapstructure:"insecure_skip_verify"`
	MinVersion         string `mapstructure:"min_version"`
	MaxVersion         string `mapstructure:"max_version"`
}

type BadgesConfig struct {
	CacheTTL *time.Duration `mapstructure:"cache_ttl"`
}

type DaemonConfig struct {
	PidFile       string `mapstructure:"pid_file"`
	LogFile       string `mapstructure:"log_file"`
	RestartPolicy string `mapstructure:"restart_policy"` // "no", "on-failure", "always"
	MaxRestarts   int    `mapstructure:"max_restarts"`   // 0 = unlimited
	RestartDelay  int    `mapstructure:"restart_delay"`  // seconds
}

type NatsConfig struct {
	MaxPayload         int32    `mapstructure:"max_payload"`
	JetStream          bool     `mapstructure:"jetstream"`
	JetStreamMaxMemory int64    `mapstructure:"jetstream_max_memory"`
	JetStreamMaxStore  int64    `mapstructure:"jetstream_max_store"`
	StoreDir           string   `mapstructure:"store_dir"`
	Routes             []string `mapstructure:"routes"`

	Cluster struct {
		Name   string `mapstructure:"name"`
		Listen string `mapstructure:"listen"`
	} `mapstructure:"cluster"`

	Log struct {
		Debug bool `mapstructure:"debug"`
		Trace bool `mapstructure:"trace"`
	} `mapstructure:"log"`
}

type PoolConfig struct {
	Connection ConnectionPoolConfig `mapstructure:"connection"`
	Workers    WorkerPoolConfig     `mapstructure:"workers"`
}

type ConnectionPoolConfig struct {
	MaxIdleTime string `mapstructure:"max_idle_time"`
	MaxSize     int    `mapstructure:"max_size"`
	ShardCount  int    `mapstructure:"shard_count"`
}

type WorkerPoolConfig struct {
	WorkerCount   int  `mapstructure:"worker_count"`
	JobQueueSize  int  `mapstructure:"job_queue_size"`
	BatchSize     int  `mapstructure:"batch_size"`
	EnableMetrics bool `mapstructure:"enable_metrics"`
}

type CatvConfig struct {
	Migration struct {
		Database orm.DatabaseConfig `mapstructure:"database"`
	} `mapstructure:"migration"`

	Database orm.DatabaseConfig `mapstructure:"database"`

	Collect struct {
		Transmission struct {
			Use bool `mapstructure:"use"`

			Daily struct {
				Port int `mapstructure:"port"`
			} `mapstructure:"daily"`

			Diagnostic struct {
				Port int `mapstructure:"port"`
			} `mapstructure:"diagnostic"`

			NetworkQualityTransition struct {
				Port int `mapstructure:"port"`
			} `mapstructure:"network_quality_transition"`

			Periodic struct {
				Port int `mapstructure:"port"`
			} `mapstructure:"periodic"`

			QualityMeasurement struct {
				Port int `mapstructure:"port"`
			} `mapstructure:"quality_measurement"`
		} `mapstructure:"transmission"`

		Weather struct {
			Use           bool   `mapstructure:"use"`
			Simulation    bool   `mapstructure:"simulation"`
			Address       string `mapstructure:"address"`
			User          string `mapstructure:"user"`
			Password      string `mapstructure:"password"`
			CronSpec      string `mapstructure:"cron_spec"`
			BaseDirectory string `mapstructure:"base_directory"`

			Timeout struct {
				Connection string `mapstructure:"connection"`
				Shut       string `mapstructure:"shut"`
			} `mapstructure:"timeout"`
		} `mapstructure:"weather"`
	} `mapstructure:"collect"`

	Control struct {
		Use                     bool `mapstructure:"use"`
		Simulation              bool `mapstructure:"simulation"`
		Port                    int  `mapstructure:"port"`
		MaxConnectionsPerSecond int  `mapstructure:"max_connections_per_second"`

		Batch struct {
			MaxCount      int    `mapstructure:"max_count"`
			MaxSize       int    `mapstructure:"max_size"`
			FlushInterval string `mapstructure:"flush_interval"`
		} `mapstructure:"batch"`

		Timeout struct {
			Connection string `mapstructure:"connection"`
			Send       string `mapstructure:"send"`
		} `mapstructure:"timeout"`

		Retry struct {
			Total struct {
				Count int    `mapstructure:"count"`
				Delay string `mapstructure:"delay"`
			} `mapstructure:"total"`

			PortExhaustion struct {
				Count int    `mapstructure:"count"`
				Delay string `mapstructure:"delay"`
			} `mapstructure:"port_exhaustion"`
		} `mapstructure:"retry"`

		K8sJob struct {
			Count                   int      `mapstructure:"count"`
			Namespace               string   `mapstructure:"namespace"`
			Image                   string   `mapstructure:"image"`
			ServiceAccountName      string   `mapstructure:"service_account_name"`
			ConfigMapName           string   `mapstructure:"config_map_name"`
			SecretName              string   `mapstructure:"secret_name"`
			ActiveDeadlineSeconds   int64    `mapstructure:"active_deadline_seconds"`
			BackoffLimit            int32    `mapstructure:"backoff_limit"`
			TTLSecondsAfterFinished int32    `mapstructure:"ttl_seconds_after_finished"`
			AllowedEnvVars          []string `mapstructure:"allowed_env_vars"`
		} `mapstructure:"k8s_job"`
	} `mapstructure:"control"`
}

type MetricsConfig struct {
	Use bool `mapstructure:"use"`
}

type ETLConfig struct {
	Elasticsearch orm.ElasticsearchConfig `mapstructure:"elasticsearch"`
}

type HomeUPMConfig struct {
	Datasource string `mapstructure:"datasource"`
}

type Config struct {
	Serve         ServeConfig              `mapstructure:"serve"`
	Servers       map[string]ServerConfig  `mapstructure:"servers"`
	Clients       map[string]ClientConfig  `mapstructure:"clients"`
	Database      orm.DatabaseConfig       `mapstructure:"database"`
	Workflow      WorkflowConfig           `mapstructure:"workflow"`
	Auth          AuthConfig               `mapstructure:"auth"`
	User          UsersConfig              `mapstructure:"users"`
	Role          map[string]CompositeRole `mapstructure:"role"` // 👈 직접 매핑
	Master        MasterConfig             `mapstructure:"master"`
	Logger        RotationConfig           `mapstructure:"logger"`
	History       HistoryConfig            `mapstructure:"history"`
	Plugins       PluginsConfig            `mapstructure:"plugins"`
	Collect       CollectConfig            `mapstructure:"collect"`
	SharedLibrary SharedLibraryConfig      `mapstructure:"shared-library"`
	ClientTLS     ClientTLSConfig          `mapstructure:"client_tls"`
	Provisioning  ProvisioningConfig       `mapstructure:"provisioning"`
	Migration     MigrationConfig          `mapstructure:"migration"`
	Badges        BadgesConfig             `mapstructure:"badges"`
	Statistics    StatisticsConfig         `mapstructure:"statistics"`
	Daemon        DaemonConfig             `mapstructure:"daemon"`
	Nats          NatsConfig               `mapstructure:"nats"`
	Pool          PoolConfig               `mapstructure:"pool"`
	Catv          CatvConfig               `mapstructure:"catv"`
	Metrics       MetricsConfig            `mapstructure:"metrics"`
	ETL           ETLConfig                `mapstructure:"etl"`
	HomeUPM       HomeUPMConfig            `mapstructure:"home_upm"`
}
