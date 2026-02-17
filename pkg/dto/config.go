package dto

import "time"

type APIConfig struct {
	Server   serverConfig   `mapstructure:"server"`
	Database databaseConfig `mapstructure:"database"`
	Nats     natsConfig     `mapstructure:"nats"`
}

type serverConfig struct {
	Port            string        `mapstructure:"port"`
	ShutdownTimeout time.Duration `mapstructure:"shutdown_timeout"`
}

type DatabaseConfig struct {
	Host     string `mapstructure:"host"`
	Port     string `mapstructure:"port"`
	User     string `mapstructure:"user"`
	Password string `mapstructure:"password"`
	DBName   string `mapstructure:"db_name"`
}

// databaseConfig is an alias for backward compatibility
type databaseConfig = DatabaseConfig

type natsConfig struct {
	URL string `mapstructure:"url"`
}
