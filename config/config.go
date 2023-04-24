package config

import (
	"github.com/caarlos0/env"
	"github.com/go-playground/validator/v10"
)

// Config contains environment values
type Config struct {
	DebugLog bool `env:"DEBUG_LOG" envDefault:"false"`
	LogJSON  bool `env:"LOG_JSON" envDefault:"true"`

	// ---server config -----
	// destination server address where traffic needs to be proxied
	ProxyTo string `env:"PROXY_TO" envDefault:"grpc.osmosis.zone:9090"`
	// proxy server listening address, by default listening at 9090
	ListenPort string `env:"LISTEN_PORT" envDefault:"9090"`

	// ---client config -----
	// ServerAddress where client to be connected
	// assume its running locally on 9090
	ServerAddress string `env:"SERVER_ADDRESS" envDefault:"localhost:9090"`
}

// LoadFromEnv parses environment variables into a given struct and validates
// its fields' values.
func LoadFromEnv(config interface{}) error {
	if err := env.Parse(config); err != nil {
		return err
	}
	if err := validator.New().Struct(config); err != nil {
		return err
	}
	return nil
}

func NewConfig() *Config {
	var cfg Config
	if err := LoadFromEnv(&cfg); err != nil {
		panic(err)
	}
	return &cfg
}
