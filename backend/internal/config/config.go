package config

import (
	"os"
	"strings"

	"gopkg.in/yaml.v2"
)

var config *Config

type option struct {
	configFile string
}

// Option ...
type Option func(*option)

// WithConfigFile ...
func WithConfigFile(file string) Option {
	return func(opt *option) {
		opt.configFile = file
	}
}

// Init membaca file yaml. Environment variable (DB_DSN, JWT_SECRET,
// CORS_ORIGINS, ADDR, APP_TZ) bila diisi akan menimpa isi yaml.
func Init(opts ...Option) error {
	opt := &option{
		configFile: getDefaultConfigFile(),
	}
	for _, optFunc := range opts {
		optFunc(opt)
	}

	out, err := os.ReadFile(opt.configFile)
	if err != nil {
		return err
	}

	cfg := &Config{}
	if err = yaml.Unmarshal(out, cfg); err != nil {
		return err
	}

	if v := os.Getenv("DB_DSN"); v != "" {
		cfg.Database.Bakery = v
	}
	if v := os.Getenv("JWT_SECRET"); v != "" {
		cfg.JWT.Secret = v
	}
	if v := os.Getenv("CORS_ORIGINS"); v != "" {
		cfg.CORS.Origins = strings.Split(v, ",")
	}
	if v := os.Getenv("ADDR"); v != "" {
		cfg.Server.Port = v
	}
	if v := os.Getenv("APP_TZ"); v != "" {
		cfg.Server.Timezone = v
	}
	if cfg.Server.Port == "" {
		cfg.Server.Port = ":8080"
	}
	if cfg.Server.Timezone == "" {
		cfg.Server.Timezone = "Asia/Jakarta"
	}

	config = cfg
	return nil
}

func getDefaultConfigFile() string {
	if v := os.Getenv("CONFIG_FILE"); v != "" {
		return v
	}
	return "./files/etc/bakery-scheduler/bakery-scheduler.development.yaml"
}

// Get ...
func Get() *Config {
	return config
}
