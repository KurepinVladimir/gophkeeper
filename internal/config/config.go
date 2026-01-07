// Package config provides application configuration loading.
// It supports configuration from files, environment variables, and command-line flags,
// applying them in the following priority order:
//
//	file < environment variables < command-line flags
//
// Configuration is loaded using Viper and mapped into strongly typed structures.
package config

import (
	"fmt"

	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

// ServerConfig defines all runtime configuration parameters of the server.
type ServerConfig struct {
	Addr         string `mapstructure:"addr"`
	DSN          string `mapstructure:"dsn"`
	JWTSecret    string `mapstructure:"jwt_secret"`
	MasterKey    string `mapstructure:"master_key"`
	TLSCert      string `mapstructure:"tls_cert"`
	TLSKey       string `mapstructure:"tls_key"`
	ReadTimeout  string `mapstructure:"read_timeout"`
	WriteTimeout string `mapstructure:"write_timeout"`
}

// Load loads application configuration into ServerConfig.
// Configuration values are resolved in the following order:
//
//  1. Configuration file (optional)
//  2. Environment variables
//  3. Command-line flags
//
// The function returns an error if required configuration fields are missing
// or if configuration cannot be unmarshaled.
func Load(flags *pflag.FlagSet) (ServerConfig, error) {
	v := viper.New()

	// Configuration file settings.
	v.SetConfigName("config")
	v.SetConfigType("json")
	v.AddConfigPath(".")
	v.AddConfigPath("./configs")

	// Default values.
	v.SetDefault("addr", "127.0.0.1:8080")
	v.SetDefault("read_timeout", "5s")
	v.SetDefault("write_timeout", "5s")

	// Environment variables.
	v.SetEnvPrefix("GK")
	v.AutomaticEnv()

	// Configuration file is optional.
	_ = v.ReadInConfig()

	// Bind command-line flags.
	_ = v.BindPFlag("addr", flags.Lookup("addr"))
	_ = v.BindPFlag("dsn", flags.Lookup("dsn"))
	_ = v.BindPFlag("jwt_secret", flags.Lookup("jwt"))
	_ = v.BindPFlag("read_timeout", flags.Lookup("read_timeout"))
	_ = v.BindPFlag("write_timeout", flags.Lookup("write_timeout"))
	_ = v.BindPFlag("master_key", flags.Lookup("master_key"))
	_ = v.BindPFlag("tls_cert", flags.Lookup("tls_cert"))
	_ = v.BindPFlag("tls_key", flags.Lookup("tls_key"))

	var cfg ServerConfig
	if err := v.Unmarshal(&cfg); err != nil {
		return ServerConfig{}, err
	}

	// Required configuration validation.
	if cfg.DSN == "" {
		return ServerConfig{}, fmt.Errorf("dsn is required")
	}
	if cfg.JWTSecret == "" {
		return ServerConfig{}, fmt.Errorf("jwt_secret is required")
	}

	return cfg, nil
}
