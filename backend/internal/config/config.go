package config

import (
	"time"
	"strings"
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
	OIDCIssuerURL    string `mapstructure:"oidc_issuer_url"`
	OIDCClientId     string `mapstructure:"oidc_client_id"`
	OIDCClientSecret string `mapstructure:"oidc_client_secret"`
	OIDCRedirectURL  string `mapstructure:"oidc_redirect_url"`
	SessionSecret    string `mapstructure:"session_secret"`
	SuperAdminEmail  string `mapstructure:"super_admin_email"`
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
	v.SetDefault("server.domain", "localhost")
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

	// Environment variables: DATABASE_DSN → database.dsn
	v.SetEnvPrefix("")
	v.AutomaticEnv()
	v.SetEnvKeyReplacer(strings.NewReplacer("_", "."))

	// Read config file (optional — env vars alone are enough)
	_ = v.ReadInConfig()

	var config Config
	if err := v.Unmarshal(&config); err != nil {
		return nil, err
	}

	return &config, nil
}


