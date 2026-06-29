package config

import (
	"fmt"
	"time"

	"github.com/spf13/viper"
)

type Config struct {
	App      AppConfig
	Server   ServerConfig
	Database DatabaseConfig
	Redis    RedisConfig
	Temporal TemporalConfig
	LLM      LLMConfig
	Milvus   MilvusConfig
	JWT      JWTConfig
	Log      LogConfig
}

type AppConfig struct {
	Name string
	Mode string
}

type ServerConfig struct {
	Port         string
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
}

type DatabaseConfig struct {
	Host     string
	Port     int
	User     string
	Password string
	DBName   string
}

type RedisConfig struct {
	Host     string
	Port     int
	Password string
	DB       int
}

type TemporalConfig struct {
	Host      string
	Port      int
	Namespace string
}

type LLMConfig struct {
	Provider    string
	Model       string
	APIKey      string
	Temperature float64
	MaxTokens   int
	BaseURL     string
	EnableRAG   bool // 是否启用RAG功能
}

type MilvusConfig struct {
	Host       string
	Port       int
	Collection string
}

type JWTConfig struct {
	Secret string
	Expire time.Duration
}

type LogConfig struct {
	Level  string
	Format string
}

func Init() error {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("./configs")
	viper.AddConfigPath("./config")
	viper.AddConfigPath("./")
	viper.AddConfigPath("../configs")
	viper.AddConfigPath("../../configs")

	viper.SetDefault("app.name", "AiOpsHub Backend")
	viper.SetDefault("app.mode", "debug")
	viper.SetDefault("server.port", "8080")
	viper.SetDefault("server.read_timeout", "10s")
	viper.SetDefault("server.write_timeout", "10s")

	viper.SetDefault("database.host", "localhost")
	viper.SetDefault("database.port", 5432)
	viper.SetDefault("database.user", "aiops")
	viper.SetDefault("database.password", "aiops123")
	viper.SetDefault("database.dbname", "aiopsdb")

	viper.SetDefault("redis.host", "localhost")
	viper.SetDefault("redis.port", 6379)
	viper.SetDefault("redis.password", "")
	viper.SetDefault("redis.db", 0)

	viper.SetDefault("temporal.host", "localhost")
	viper.SetDefault("temporal.port", 7233)
	viper.SetDefault("temporal.namespace", "default")

	viper.SetDefault("llm.provider", "openai")
	viper.SetDefault("llm.model", "gpt-4")
	viper.SetDefault("llm.temperature", 0.7)
	viper.SetDefault("llm.max_tokens", 2000)
	viper.SetDefault("llm.enable_rag", true) // 默认启用RAG

	viper.SetDefault("jwt.secret", "aiops-jwt-secret-key-2024")
	viper.SetDefault("jwt.expire", "24h")

	viper.SetDefault("log.level", "info")
	viper.SetDefault("log.format", "console")

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			fmt.Println("Config file not found, using defaults")
		} else {
			return err
		}
	} else {
		fmt.Println("Using config file:", viper.ConfigFileUsed())
	}

	return nil
}

func GetConfig() *Config {
	return &Config{
		App: AppConfig{
			Name: viper.GetString("app.name"),
			Mode: viper.GetString("app.mode"),
		},
		Server: ServerConfig{
			Port:         viper.GetString("server.port"),
			ReadTimeout:  viper.GetDuration("server.read_timeout"),
			WriteTimeout: viper.GetDuration("server.write_timeout"),
		},
		Database: DatabaseConfig{
			Host:     viper.GetString("database.host"),
			Port:     viper.GetInt("database.port"),
			User:     viper.GetString("database.user"),
			Password: viper.GetString("database.password"),
			DBName:   viper.GetString("database.dbname"),
		},
		Redis: RedisConfig{
			Host:     viper.GetString("redis.host"),
			Port:     viper.GetInt("redis.port"),
			Password: viper.GetString("redis.password"),
			DB:       viper.GetInt("redis.db"),
		},
		Temporal: TemporalConfig{
			Host:      viper.GetString("temporal.host"),
			Port:      viper.GetInt("temporal.port"),
			Namespace: viper.GetString("temporal.namespace"),
		},
		LLM: LLMConfig{
			Provider:    viper.GetString("llm.provider"),
			Model:       viper.GetString("llm.model"),
			APIKey:      viper.GetString("llm.api_key"),
			Temperature: viper.GetFloat64("llm.temperature"),
			MaxTokens:   viper.GetInt("llm.max_tokens"),
			BaseURL:     viper.GetString("llm.base_url"),
			EnableRAG:   viper.GetBool("llm.enable_rag"),
		},
		Milvus: MilvusConfig{
			Host:       viper.GetString("milvus.host"),
			Port:       viper.GetInt("milvus.port"),
			Collection: viper.GetString("milvus.collection"),
		},
		JWT: JWTConfig{
			Secret: viper.GetString("jwt.secret"),
			Expire: viper.GetDuration("jwt.expire"),
		},
		Log: LogConfig{
			Level:  viper.GetString("log.level"),
			Format: viper.GetString("log.format"),
		},
	}
}
