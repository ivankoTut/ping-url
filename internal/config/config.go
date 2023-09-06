package config

import (
	"github.com/ilyakaznacheev/cleanenv"
	"log"
	"os"
)

type (
	// Config структура повторяет данные yaml конфига
	Config struct {
		Env             string   `yaml:"env" env-default:"prod"`
		LogFile         string   `yaml:"log_file" env-default:"prod.log"`
		BotToken        string   `yaml:"bot_token" env-required:"true"`
		Database        Database `yaml:"database" env-required:"true"`
		Jaeger          Jaeger   `yaml:"jaeger" env-required:"true"`
		DefaultTimePing int64    `yaml:"default_time_ping" env-default:"300"`
		AccessUserList  []int64  `yaml:"access_user_list"`
	}

	Database struct {
		Clickhouse Clickhouse `yaml:"clickhouse" env-required:"true"`
		Postgres   Postgres   `yaml:"postgres" env-required:"true"`
		Redis      Redis      `yaml:"redis" env-required:"true"`
	}

	Postgres struct {
		Database      string `yaml:"database" env-required:"true"`
		User          string `yaml:"user" env-required:"true"`
		Password      string `yaml:"password" env-required:"true"`
		Host          string `yaml:"host" env-required:"true"`
		Port          uint64 `yaml:"port" env-required:"true"`
		MigrationPath string `yaml:"migration_path" env-required:"true"`
	}

	Clickhouse struct {
		Database      string `yaml:"database" env-required:"true"`
		User          string `yaml:"user" env-required:"true"`
		Password      string `yaml:"password" env-required:"true"`
		Host          string `yaml:"host" env-required:"true"`
		Port          uint64 `yaml:"port" env-required:"true"`
		MigrationPath string `yaml:"migration_path" env-required:"true"`
	}

	Redis struct {
		Addr     string `yaml:"addr" env-required:"true"`
		Password string `yaml:"password"`
		Db       int    `yaml:"db" env-default:"0"`
	}

	Jaeger struct {
		Url  string `yaml:"url" env-required:"true"`
		Name string `yaml:"name" env-required:"true"`
		Env  string `yaml:"env" env-default:"production"`
	}
)

func MustLoadConfig() *Config {
	configPath := os.Getenv("CONFIG_PATH")
	if configPath == "" {
		log.Fatal("CONFIG_PATH is not set")
	}

	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		log.Fatalf("config file does not exist: %s", configPath)
	}

	var cfg Config

	if err := cleanenv.ReadConfig(configPath, &cfg); err != nil {
		log.Fatalf("cannot read config: %s, error: %s", configPath, err)
	}

	return &cfg
}
