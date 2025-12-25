package config

import "github.com/caarlos0/env/v11"

type Config struct {
	DB      DBConfig
	Proxmox ProxmoxConfig
	Traq    TraqConfig
}

type DBConfig struct {
	User     string `env:"NS_MARIADB_USER" envDefault:"user"`
	Password string `env:"NS_MARIADB_PASSWORD" envDefault:"password"`
	Host     string `env:"NS_MARIADB_HOST" envDefault:"localhost"`
	Port     int    `env:"NS_MARIADB_PORT" envDefault:"3306"`
	Name     string `env:"NS_MARIADB_DATABASE" envDefault:"database"`
}

type ProxmoxConfig struct {
	Endpoint string `env:"PROXMOX_ENDPOINT,notEmpty"`
	TokenID  string `env:"PROXMOX_TOKEN_ID,notEmpty"`
	Secret   string `env:"PROXMOX_SECRET,notEmpty"`
}

type TraqConfig struct {
	Endpoint string `env:"TRAQ_ENDPOINT,notEmpty"`
	ApiToken string `env:"TRAQ_API_TOKEN,notEmpty"`
}

func LoadConfig() (*Config, error) {
	var cfg Config
	if err := env.Parse(&cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}
