package config

import (
	"flag"
	"os"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
	"github.com/joho/godotenv"
)

type Config struct {
	Env      string         `yaml:"env"`
	Postgres PostgresConfig `yaml:"postgres"`
	HTTP     ServerConfig   `yaml:"http"`
}

type PostgresConfig struct {
	Host          string `env:"POSTGRES_HOST"`
	Port          string `env:"POSTGRES_PORT"`
	DbName        string `env:"POSTGRES_DB"`
	User          string `env:"POSTGRES_USER"`
	Password      string `env:"POSTGRES_PASSWORD_FILE"`
	Pool          PoolConfig
	MigrationsDir string `yaml:"migrations_dir"`
}

type PoolConfig struct {
	MaxConns        int    `yaml:"max_conns"`
	MaxConnLifetime string `yaml:"max_conn_lifetime"`
	ConnectTimeout  string `yaml:"connect_timeout"`
}

type ServerConfig struct {
	Addr         string        `yaml:"addr"`
	ReadTimeout  time.Duration `yaml:"read_timeout"`
	WriteTimeout time.Duration `yaml:"write_timeout"`
	IdleTimeout  time.Duration `yaml:"idle_timeout"`
}

func MustLoad() *Config {
	configPath := fetchConfigPath()
	if configPath == "" {
		panic("config path is empty")
	}

	return MustLoadPath(configPath)
}

func MustLoadPath(configPath string) *Config {
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		panic("config file does not exist: " + configPath)
	}

	var cfg Config

	if err := cleanenv.ReadConfig(configPath, &cfg); err != nil {
		panic("cannot read config: " + err.Error())
	}

	if cfg.Env == "" {
		panic("env var cannot be empty")
	}

	err := setEnvValues()
	if err != nil {
		panic(err)
	}

	err = cleanenv.ReadEnv(&cfg)
	if err != nil {
		panic("cannot read env vars: " + err.Error())
	}
	loadPassword(&cfg)

	return &cfg
}

func loadPassword(cfg *Config) {
	if cfg.Postgres.Password == "" {
		panic("password is empty")
	}

	data, err := os.ReadFile(cfg.Postgres.Password)
	if err != nil {
		panic("cannot read file with password")
	}
	cfg.Postgres.Password = string(data)
}

func setEnvValues() error {
	if err := godotenv.Load(); err != nil {
		panic("cannot load local .env")
	}

	return nil
}

// Priority: flag > env > default
func fetchConfigPath() string {
	var res string

	flag.StringVar(&res, "config", "", "path to config file")
	flag.Parse()

	if res == "" {
		res = os.Getenv("CONFIG_PATH")
	}

	return res
}
