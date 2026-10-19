package config

type (
	// Config ...
	Config struct {
		Server   ServerConfig   `yaml:"server"`
		Database DatabaseConfig `yaml:"database"`
		JWT      JWTConfig      `yaml:"jwt"`
		CORS     CORSConfig     `yaml:"cors"`
	}

	// ServerConfig ...
	ServerConfig struct {
		Port     string `yaml:"port"`
		Timezone string `yaml:"timezone"`
	}

	// DatabaseConfig ...
	DatabaseConfig struct {
		Bakery string `yaml:"bakery"`
	}

	// JWTConfig ...
	JWTConfig struct {
		Secret string `yaml:"secret"`
	}

	// CORSConfig ...
	CORSConfig struct {
		Origins []string `yaml:"origins"`
	}
)
