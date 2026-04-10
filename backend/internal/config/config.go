package config

import (
	"time"

	"github.com/spf13/viper"
)

// Config holds all application configuration.
type Config struct {
	Server    ServerConfig    `mapstructure:"server"`
	Database  DatabaseConfig  `mapstructure:"database"`
	FileStore FileStoreConfig `mapstructure:"filestore"`
	Cache     CacheConfig     `mapstructure:"cache"`
	Auth      AuthConfig      `mapstructure:"auth"`
	Version   VersionConfig   `mapstructure:"versioning"`
}

type ServerConfig struct {
	Domain string `mapstructure:"domain"`
	Port   int    `mapstructure:"port"`
}

type DatabaseConfig struct {
	Driver         string `mapstructure:"driver"`
	DSN            string `mapstructure:"dsn"`
	MaxOpenConns   int    `mapstructure:"max_open_conns"`
	MaxIdleConns   int    `mapstructure:"max_idle_conns"`
	ConnMaxLifetime time.Duration `mapstructure:"conn_max_lifetime"`
}

type FileStoreConfig struct {
	Driver   string `mapstructure:"driver"`
	Endpoint string `mapstructure:"endpoint"`
	Bucket   string `mapstructure:"bucket"`
	Region   string `mapstructure:"region"`
}

type CacheConfig struct {
	Driver       string        `mapstructure:"driver"`
	Address      string        `mapstructure:"addr"`
	TimeToLive   time.Duration `mapstructure:"ttl"`
	MaxCostBytes int64         `mapstructure:"max_cost_bytes"`
}

type AuthConfig struct {
	OIDCIssuerURL       string `mapstructure:"oidc_issuer_url"`
	OIDCClientId        string `mapstructure:"oidc_client_id"`
	OIDCClientSecret    string `mapstructure:"oidc_client_secret"`
	OIDCRedirectURL     string `mapstructure:"oidc_redirect_url"`
	SessionSecret       string `mapstructure:"session_secret"`
	SuperAdminEmail     string `mapstructure:"super_admin_email"`
	SuperAdminAPIKey    string `mapstructure:"super_admin_api_key"`
	SuperAdminName      string `mapstructure:"super_admin_display_name"`
}

type VersionConfig struct {
	InactivityWindow   time.Duration `mapstructure:"inactivity_window"`
	CommitPollInterval time.Duration `mapstructure:"commit_poll_interval"`
}

// Load reads configuration from a file and environment variables.
// Environment variables override file values (e.g. DATABASE_DSN overrides database.dsn).
func Load(configPath string) (*Config, error) {
	v := viper.New()

	// Defaults
	v.SetDefault("server.domain", "0.0.0.0")
	v.SetDefault("server.port", 8088)
	v.SetDefault("database.driver", "postgres")
	v.SetDefault("database.max_open_conns", 25)
	v.SetDefault("database.max_idle_conns", 5)
	v.SetDefault("database.conn_max_lifetime", 5*time.Minute)
	v.SetDefault("filestore.driver", "s3")
	v.SetDefault("filestore.region", "us-east-1")
	v.SetDefault("cache.driver", "memory")
	v.SetDefault("cache.ttl", 5*time.Minute)
	v.SetDefault("cache.max_cost_bytes", 64*1024*1024) // 64 MB
	v.SetDefault("auth.super_admin_display_name", "Super Admin")
	v.SetDefault("versioning.inactivity_window", 5*time.Minute)
	v.SetDefault("versioning.commit_poll_interval", 60*time.Second)

	// File
	if configPath != "" {
		v.SetConfigFile(configPath)
	} else {
		v.SetConfigName("config")
		v.SetConfigType("yaml")
		v.AddConfigPath(".")
		v.AddConfigPath("/app")
	}

	// Explicitly bind env vars to config keys.
	// AutomaticEnv does not reliably resolve nested keys (e.g. database.dsn),
	// so each mapping is declared here.
	_ = v.BindEnv("server.domain", "SERVER_DOMAIN")
	_ = v.BindEnv("server.port", "SERVER_PORT")
	_ = v.BindEnv("database.dsn", "DATABASE_DSN")
	_ = v.BindEnv("database.max_open_conns", "DATABASE_MAX_OPEN_CONNS")
	_ = v.BindEnv("database.max_idle_conns", "DATABASE_MAX_IDLE_CONNS")
	_ = v.BindEnv("database.conn_max_lifetime", "DATABASE_CONN_MAX_LIFETIME")
	_ = v.BindEnv("filestore.endpoint", "FILESTORE_ENDPOINT")
	_ = v.BindEnv("filestore.bucket", "FILESTORE_BUCKET")
	_ = v.BindEnv("filestore.region", "FILESTORE_REGION")
	_ = v.BindEnv("cache.ttl", "CACHE_TTL")
	_ = v.BindEnv("cache.max_cost_bytes", "CACHE_MAX_COST_BYTES")
	_ = v.BindEnv("auth.oidc_issuer_url", "AUTH_OIDC_ISSUER_URL")
	_ = v.BindEnv("auth.oidc_client_id", "AUTH_OIDC_CLIENT_ID")
	_ = v.BindEnv("auth.oidc_client_secret", "AUTH_OIDC_CLIENT_SECRET")
	_ = v.BindEnv("auth.oidc_redirect_url", "AUTH_OIDC_REDIRECT_URL")
	_ = v.BindEnv("auth.session_secret", "AUTH_SESSION_SECRET")
	_ = v.BindEnv("auth.super_admin_email", "AUTH_SUPER_ADMIN_EMAIL")
	_ = v.BindEnv("auth.super_admin_api_key", "AUTH_SUPER_ADMIN_API_KEY")
	_ = v.BindEnv("auth.super_admin_display_name", "AUTH_SUPER_ADMIN_DISPLAY_NAME")
	_ = v.BindEnv("versioning.inactivity_window", "VERSIONING_INACTIVITY_WINDOW")
	_ = v.BindEnv("versioning.commit_poll_interval", "VERSIONING_COMMIT_POLL_INTERVAL")

	// Read config file (optional — env vars alone are enough)
	_ = v.ReadInConfig()

	var config Config
	if err := v.Unmarshal(&config); err != nil {
		return nil, err
	}

	return &config, nil
}


