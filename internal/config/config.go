package config

import (
	"log"
	"os"
	"strings"
	"time"

	"github.com/begenov/real-estate/internal/logger"
	"github.com/joho/godotenv"
	"github.com/spf13/viper"
)

const (
	EnvLocal = "local"
	Prod     = "prod"
)

type Config struct {
	Postgres    DBConfig           `mapstructure:"postgres"`
	HTTP        HTTPConfig         `mapstructure:"http"`
	JWT         JWTConfig          `mapstructure:"jwt"`
	Limiter     LimiterConfig      `mapstructure:"limiter"`
	Cache       CacheConfig        `mapstructure:"cache"`
	Redis       RedisConfig        `mapstructure:"redis"`
	Minio       MinioConfig        `mapstructure:"minio"`
	SMTP        SMTPConfig         `mapstructure:"smtp"`
	EmailConfig EmailConfig        `mapstructure:"email"`
	Watermark   WatermarkConfig    `mapstructure:"watermark"`
	Exchange    ExchangeRateConfig `mapstructure:"exchange_rate"`
	Async       AsyncConfig        `mapstructure:"async"`
	Translate   TranslateConfig    `mapstructure:"translate"`
	Environment string
}

type DBConfig struct {
	Driver string `mapstructure:"DB_DRIVER"`
	DSN    string `mapstructure:"DB_SOURCE"`
}

type HTTPConfig struct {
	Addr               string        `mapstructure:"HTTP_PORT"`
	ReadTimeout        time.Duration `mapstructure:"HTTP_READ_TIMEOUT"`
	WriteTimeout       time.Duration `mapstructure:"HTTP_WRITE_TIMEOUT"`
	MaxHeaderBytes     int           `mapstructure:"HTTP_MAX_HEADER_BYTES"`
	CORSAllowedOrigins []string      `mapstructure:"CORS_ALLOWED_ORIGINS"`
}

type JWTConfig struct {
	AccessTokenTTL    time.Duration `mapstructure:"AUTH_ACCESS_TOKEN_TTL"`
	RefreshTokenTTL   time.Duration `mapstructure:"AUTH_REFRESH_TOKEN_TTL"`
	AccessSigningKey  string        `mapstructure:"AUTH_ACCESS_SIGNING_KEY"`
	RefreshSigningKey string        `mapstructure:"AUTH_REFRESH_SIGNING_KEY"`
}

type LimiterConfig struct {
	RPS   int           `mapstructure:"LIMITER_RPS"`
	Burst int           `mapstructure:"LIMITER_BURST"`
	TTL   time.Duration `mapstructure:"LIMITER_TTL"`
}

type RedisConfig struct {
	Host     string `mapstructure:"REDIS_HOST"`
	Port     string `mapstructure:"REDIS_PORT"`
	Username string `mapstructure:"REDIS_USERNAME"`
	Password string `mapstructure:"REDIS_PASSWORD"`
	DB       int    `mapstructure:"REDIS_DB"`
}

type MinioConfig struct {
	Endpoint  string   `mapstructure:"MINIO_ENDPOINT"`
	AccessKey string   `mapstructure:"MINIO_ACCESS_KEY"`
	SecretKey string   `mapstructure:"MINIO_SECRET_KEY"`
	Buckets   []string `mapstructure:"MINIO_BUCKET_NAME"`
	Secure    bool     `mapstructure:"MINIO_SECURE"`
}

type SMTPConfig struct {
	Host     string `mapstructure:"SMTP_HOST"`
	Port     int    `mapstructure:"SMTP_PORT"`
	From     string `mapstructure:"SMTP_FROM"`
	Password string `mapstructure:"SMTP_PASSWORD"`
}

type EmailConfig struct {
	Templates EmailTemplate
	Subjects  EmailSubject
}

type EmailTemplate struct {
	ContractFormEmail string `mapstructure:"TEMPLATE_CONTACT_FORM_EMAIL"`
	TourFormEmail     string `mapstructure:"TEMPLATE_TOUR_FORM_EMAIL"`
}

type EmailSubject struct {
	ContractFormEmail string `mapstructure:"SUBJECT_CONTACT_FORM_EMAIL"`
	TourFormEmail     string `mapstructure:"SUBJECT_TOUR_FORM_EMAIL"`
}

type CacheConfig struct {
	TTL time.Duration `mapstructure:"CACHE_TTL"`
}

type WatermarkConfig struct {
	Path string `mapstructure:"WATERMARK_PATH"`
}

type ExchangeRateConfig struct {
	APIKey string `mapstructure:"EXCHANGE_API_KEY"`
}

type AsyncConfig struct {
	TranslateConcurrency  int `mapstructure:"TRANSLATE_CONCURRENCY"`
	MinioReprocessWorkers int `mapstructure:"MINIO_REPROCESS_WORKERS"`
	MinioReprocessQueue   int `mapstructure:"MINIO_REPROCESS_QUEUE"`
}

type TranslateConfig struct {
	CredentialsPath string `mapstructure:"GOOGLE_APPLICATION_CREDENTIALS"`
}

func NewConfig(configDir, envName, configType string) (*Config, error) {
	loadEnvFile(configDir + "/" + envName + "." + configType)

	viper.AutomaticEnv()

	viper.AddConfigPath(configDir)
	viper.SetConfigName(envName)
	viper.SetConfigType(configType)

	if err := viper.ReadInConfig(); err != nil {
		logger.Warn("No .env file found, using system ENV")
	}

	var cfg Config
	if err := unmarshal(&cfg); err != nil {
		logger.Error("unmarshal(): ", err)
		return nil, err
	}

	return &cfg, nil
}

func unmarshal(cfg *Config) error {
	cfg.Environment = os.Getenv("APP_ENV")

	if err := viper.Unmarshal(&cfg.Postgres); err != nil {
		return err
	}

	if err := viper.Unmarshal(&cfg.HTTP); err != nil {
		return err
	}

	if err := viper.Unmarshal(&cfg.JWT); err != nil {
		return err
	}

	if err := viper.Unmarshal(&cfg.Limiter); err != nil {
		return err
	}

	if err := viper.Unmarshal(&cfg.Cache); err != nil {
		return err
	}

	if err := viper.Unmarshal(&cfg.Redis); err != nil {
		return err
	}

	if err := viper.Unmarshal(&cfg.Minio); err != nil {
		return err
	}

	if err := viper.Unmarshal(&cfg.SMTP); err != nil {
		return err
	}

	if err := viper.Unmarshal(&cfg.EmailConfig.Subjects); err != nil {
		return err
	}

	if err := viper.Unmarshal(&cfg.EmailConfig.Templates); err != nil {
		return err
	}

	if err := viper.Unmarshal(&cfg.Watermark); err != nil {
		return err
	}

	if err := viper.Unmarshal(&cfg.Exchange); err != nil {
		return err
	}
	if err := viper.Unmarshal(&cfg.Async); err != nil {
		return err
	}
	if err := viper.Unmarshal(&cfg.Translate); err != nil {
		return err
	}

	cfg.Minio.Buckets = splitAndTrim(viper.GetString("MINIO_BUCKETS"))
	cfg.HTTP.CORSAllowedOrigins = splitAndTrim(viper.GetString("CORS_ALLOWED_ORIGINS"))

	if cfg.Async.TranslateConcurrency == 0 {
		cfg.Async.TranslateConcurrency = 3
	}
	if cfg.Async.MinioReprocessWorkers == 0 {
		cfg.Async.MinioReprocessWorkers = 4
	}
	if cfg.Async.MinioReprocessQueue == 0 {
		cfg.Async.MinioReprocessQueue = 100
	}

	return nil
}

func splitAndTrim(input string) []string {
	parts := strings.Split(input, ",")
	result := make([]string, 0, len(parts))
	for _, part := range parts {
		value := strings.TrimSpace(part)
		if value == "" {
			continue
		}
		result = append(result, value)
	}
	return result
}

func loadEnvFile(path string) {
	if err := godotenv.Load(path); err != nil {
		log.Println("No .env file found, using system ENV")
	}
}
