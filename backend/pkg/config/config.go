package config

import (
	"log"
	"os"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	Env        string `yaml:"env" env-default:"local"`
	Database   `yaml:"database"`
	JWT `yaml:"jwt"`
	HTTPServer `yaml:"http_server"`
	Connections `yaml:"connections"`
}

type Database struct {
	User     string `yaml:"user"`
	Password string `yaml:"password"`
	Port     string `yaml:"port"`
	Host     string `yaml:"host"`
	DB       string `yaml:"db"`
}

type JWT struct {
	JWTKey   string        `yaml:"jwt_key"`
	TokenTTL time.Duration `yaml:"token_ttl"`
	ShaSalt  string        `yaml:"salt"`
}

type HTTPServer struct {
	Address     string        `yaml:"address" env-default:"localhost:8080"`
	Timeout     time.Duration `yaml:"timeout" env-default:"4s"`
	IdleTimeout time.Duration `yaml:"idle_timeout" env-default:"60s"`
}
type Connections struct{
	Mlconnection string `yaml:"ml_connection"`
	ParserConnection string `yaml:"parser_connection"`
}

func LoadConfig(configPath string) *Config {
	if configPath == "" {
		log.Fatal("config path is null")
	}

	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		log.Fatalf("config file does not exist: %s", configPath)
	}

	var cfg Config

	if err := cleanenv.ReadConfig(configPath, &cfg); err != nil {
		log.Fatalf("cannot read config: %s", err)
	}

	return &cfg
}
