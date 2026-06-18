package config

import (
	"time"

	"github.com/jacky-htg/go-libs/env"
)

type Config struct {
	Server   ServerConfig
	Database DatabaseConfig
	Cache    CacheConfig
	Token    TokenConfig
	TTL      TTLConfig
}

type ServerConfig struct {
	AppPort                 int
	WriteTimeout            time.Duration
	ReadTimeout             time.Duration
	IdleTimeout             time.Duration
	GracefulShutdownTimeout time.Duration
	GatewayTimeout          time.Duration
}

type DatabaseConfig struct {
	Host            string
	Port            string
	Username        string
	Password        string
	Database        string
	SslMode         string
	Schema          string
	ApplicationName string
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration
	ConnMaxIdleTime time.Duration
}

type CacheConfig struct {
	Host     string
	Port     int
	Username string
	Password string

	// Connection Pool
	MaxIdleConns    int
	MaxActiveConns  int
	ConnMaxLifetime time.Duration
	ConnMaxIdleTime time.Duration

	// Timeout
	DialTimeout  time.Duration
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
	PoolTimeout  time.Duration

	ClientName string
	DatabaseId int

	// TLS/SSL
	TLSEnabled  bool
	TLSCertFile string
	TLSKeyFile  string
	TLSCAFile   string

	// Cluster/High Availability
	ClusterMode  bool
	ClusterNodes []string
	ReadFrom     string

	// Retry & Fallback
	MaxRetries      int
	MinRetryBackoff time.Duration
	MaxRetryBackoff time.Duration
}

type TokenConfig struct {
	TokenSalt string
	TokenExp  int
}

type TTLConfig struct {
	TTLDefault         time.Duration
	TTLHourLong        time.Duration
	TTLHourMedium      time.Duration
	TTLHourShort       time.Duration
	TTLHourVeryShort   time.Duration
	TTLMinuteLong      time.Duration
	TTLMinuteMedium    time.Duration
	TTLMinuteShort     time.Duration
	TTLMinuteVeryShort time.Duration
}

func LoadConfig() (Config, error) {
	err := env.InitEnv()
	if err != nil {
		return Config{}, err
	}

	server := ServerConfig{
		AppPort:                 env.EnvInt("APP_PORT", 9000),
		WriteTimeout:            env.EnvDuration("SERVER_WRITE_TIMEOUT", 5*time.Second),
		ReadTimeout:             env.EnvDuration("SERVER_READ_TIMEOUT", 5*time.Second),
		IdleTimeout:             env.EnvDuration("SERVER_IDLE_TIMEOUT", 30*time.Second),
		GracefulShutdownTimeout: env.EnvDuration("SERVER_GRACEFUL_SHUTDOWN_TIMEOUT", 30*time.Second),
		GatewayTimeout:          env.EnvDuration("SERVER_GATEWAY_TIMEOUT", 5*time.Second),
	}

	tokenConfig := TokenConfig{
		TokenSalt: env.Env("TOKEN_SALT", ""),
		TokenExp:  env.EnvInt("TOKEN_EXP", 5),
	}

	cacheConfig := CacheConfig{
		// Koneksi Dasar
		Host:       env.Env("VALKEY_HOST", "localhost"),
		Port:       env.EnvInt("VALKEY_PORT", 6379),
		Username:   env.Env("VALKEY_USERNAME", ""),
		Password:   env.Env("VALKEY_PASSWORD", ""),
		ClientName: env.Env("APP_NAME", ""),
		DatabaseId: env.EnvInt("VALKEY_DB_ID", 0),

		// Connection Pool - dioptimasi untuk production
		MaxIdleConns:    env.EnvInt("VALKEY_MAX_IDLE_CONNS", 10),
		MaxActiveConns:  env.EnvInt("VALKEY_MAX_ACTIVE_CONNS", 50),
		ConnMaxLifetime: env.EnvDuration("VALKEY_CONN_MAX_LIFETIME", 1*time.Hour),
		ConnMaxIdleTime: env.EnvDuration("VALKEY_CONN_MAX_IDLE_TIME", 10*time.Minute),

		// Timeout - cukup ketat untuk mencegah hanging
		DialTimeout:  env.EnvDuration("VALKEY_DIAL_TIMEOUT", 5*time.Second),
		ReadTimeout:  env.EnvDuration("VALKEY_READ_TIMEOUT", 3*time.Second),
		WriteTimeout: env.EnvDuration("VALKEY_WRITE_TIMEOUT", 3*time.Second),
		PoolTimeout:  env.EnvDuration("VALKEY_POOL_TIMEOUT", 4*time.Second),

		// TLS - aktif di environment cloud
		TLSEnabled:  env.EnvBool("VALKEY_TLS_ENABLED", false),
		TLSCertFile: env.Env("VALKEY_TLS_CERT_FILE", ""),
		TLSKeyFile:  env.Env("VALKEY_TLS_KEY_FILE", ""),
		TLSCAFile:   env.Env("VALKEY_TLS_CA_FILE", ""),

		// Cluster - jika menggunakan mode cluster
		ClusterMode:  env.EnvBool("VALKEY_CLUSTER_MODE", false),
		ClusterNodes: env.EnvSliceString("VALKEY_CLUSTER_NODES", []string{}),
		ReadFrom:     env.Env("VALKEY_READ_FROM", "MASTER"),

		// Retry - untuk resilience
		MaxRetries:      env.EnvInt("VALKEY_MAX_RETRIES", 3),
		MinRetryBackoff: env.EnvDuration("VALKEY_MIN_RETRY_BACKOFF", 8*time.Millisecond),
		MaxRetryBackoff: env.EnvDuration("VALKEY_MAX_RETRY_BACKOFF", 512*time.Millisecond),
	}

	databaseConfig := DatabaseConfig{
		Host:            env.Env("DB_HOST", "localhost"),
		Port:            env.Env("DB_PORT", "5432"),
		Username:        env.Env("DB_USERNAME", "postgres"),
		Password:        env.Env("DB_PASSWORD", "1234"),
		Database:        env.Env("DB_DATABASE", "workshop"),
		SslMode:         env.Env("DB_SSLMODE", "disable"),
		Schema:          env.Env("DB_SCHEMA", "public"),
		ApplicationName: env.Env("DB_APPLICATION_NAME", "workshop"),
		MaxOpenConns:    env.EnvInt("DB_MAX_OPEN_CONNS", 25),
		MaxIdleConns:    env.EnvInt("DB_MAX_IDLE_CONNS", 25),
		ConnMaxLifetime: env.EnvDuration("DB_CONN_MAX_LIFETIME", 5*time.Minute),
		ConnMaxIdleTime: env.EnvDuration("DB_CONN_MAX_IDLE_TIME", 5*time.Minute),
	}

	ttlConfig := TTLConfig{
		TTLDefault:         env.EnvDuration("TTL_DEFAULT", 24*time.Hour),
		TTLHourLong:        env.EnvDuration("TTL_HOUR_LONG", 24*time.Hour),
		TTLHourMedium:      env.EnvDuration("TTL_HOUR_MEDIUM", 12*time.Hour),
		TTLHourShort:       env.EnvDuration("TTL_HOUR_SHORT", 6*time.Hour),
		TTLHourVeryShort:   env.EnvDuration("TTL_HOUR_VERY_SHORT", 1*time.Hour),
		TTLMinuteLong:      env.EnvDuration("TTL_MINUTE_LONG", 30*time.Minute),
		TTLMinuteMedium:    env.EnvDuration("TTL_MINUTE_MEDIUM", 10*time.Minute),
		TTLMinuteShort:     env.EnvDuration("TTL_MINUTE_SHORT", 5*time.Minute),
		TTLMinuteVeryShort: env.EnvDuration("TTL_MINUTE_VERY_SHORT", 1*time.Minute),
	}

	return Config{
		Server:   server,
		Database: databaseConfig,
		Token:    tokenConfig,
		Cache:    cacheConfig,
		TTL:      ttlConfig,
	}, nil
}
