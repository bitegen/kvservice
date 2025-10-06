package config

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/joho/godotenv"
	"gopkg.in/yaml.v3"
)

const (
	defaultEnvPath  = "../.env"
	defaultYAMLPath = "../config.yaml"
)

type Config struct {
	Postgres PostgresConfig `yaml:"postgres"`
	HTTP     ServerConfig   `yaml:"http"`
}

type PostgresConfig struct {
	Host          string
	Port          string
	DbName        string
	User          string
	Password      string
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

func mustEnv(key string) (string, error) {
	val, ok := os.LookupEnv(key)
	if !ok || strings.TrimSpace(val) == "" {
		return "", fmt.Errorf("env var %s is required", key)
	}
	return val, nil
}

func loadEnv(envPath string) (PostgresConfig, error) {
	var nullCfg PostgresConfig
	if envPath == "" {
		envPath = defaultEnvPath
	}
	if err := godotenv.Load(envPath); err != nil {
		return PostgresConfig{}, fmt.Errorf(".env not found: %w", err)
	}

	host, err := mustEnv("POSTGRES_HOST")
	if err != nil {
		return nullCfg, err
	}
	port, err := mustEnv("POSTGRES_PORT")
	if err != nil {
		return nullCfg, err
	}
	db, err := mustEnv("POSTGRES_DB")
	if err != nil {
		return nullCfg, err
	}
	user, err := mustEnv("POSTGRES_USER")
	if err != nil {
		return nullCfg, err
	}
	pwFile, err := mustEnv("POSTGRES_PASSWORD_FILE")
	if err != nil {
		return nullCfg, err
	}
	pw, err := os.ReadFile(pwFile)
	if err != nil {
		return PostgresConfig{}, fmt.Errorf("read password file %s: %w", pwFile, err)
	}
	password := strings.TrimSpace(string(pw))
	if password == "" {
		return PostgresConfig{}, fmt.Errorf("password file %s is empty", pwFile)
	}

	return PostgresConfig{
		Host: host, Port: port, DbName: db,
		User: user, Password: password,
	}, nil
}

func loadYAML(yamlPath string) (PostgresConfig, ServerConfig, error) {
	if yamlPath == "" {
		yamlPath = defaultYAMLPath
	}
	f, err := os.Open(yamlPath)
	if err != nil {
		return PostgresConfig{}, ServerConfig{}, fmt.Errorf("open yaml: %w", err)
	}
	defer f.Close()

	var raw Config
	if err := yaml.NewDecoder(f).Decode(&raw); err != nil {
		return PostgresConfig{}, ServerConfig{}, fmt.Errorf("decode yaml: %w", err)
	}
	return raw.Postgres, raw.HTTP, nil
}

func LoadConfig(envPath, yamlPath string) (Config, error) {
	pg, err := loadEnv(envPath)
	if err != nil {
		return Config{}, err
	}
	postgresCfg, httpCfg, err := loadYAML(yamlPath)
	if err != nil {
		return Config{}, err
	}
	pg.Pool = postgresCfg.Pool
	pg.MigrationsDir = postgresCfg.MigrationsDir

	return Config{Postgres: pg, HTTP: httpCfg}, nil
}
