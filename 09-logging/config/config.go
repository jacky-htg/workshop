package config

import (
	"time"

	"github.com/jacky-htg/go-libs/env"
)

type Config struct {
	Server   ServerConfig
	Database DatabaseConfig
}

type ServerConfig struct {
	AppPort int

	WriteTimeout            time.Duration
	ReadTimeout             time.Duration
	IdleTimeout             time.Duration
	GracefulShutdownTimeout time.Duration
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

	return Config{
		Server:   server,
		Database: databaseConfig,
	}, nil
}
