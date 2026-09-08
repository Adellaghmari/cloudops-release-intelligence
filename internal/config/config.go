package config

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	Env            string
	HTTPAddr       string
	LogLevel       string
	Version        string
	ServiceName    string
	CORSOrigins    []string
	ShutdownWait   time.Duration
	SeedLocalData  bool
	StoreDriver    string
	DynamoTable    string
	DynamoEndpoint string
	AWSRegion      string
}

func Load() (Config, error) {
	_ = loadDotEnv(".env")
	cfg := Config{
		Env:            getenv("APP_ENV", "local"),
		HTTPAddr:       getenv("APP_HTTP_ADDR", ":8080"),
		LogLevel:       strings.ToLower(getenv("APP_LOG_LEVEL", "info")),
		Version:        getenv("APP_VERSION", "0.1.0-dev"),
		ServiceName:    getenv("APP_SERVICE_NAME", "cloudops-api"),
		ShutdownWait:   10 * time.Second,
		SeedLocalData:  true,
		StoreDriver:    getenv("APP_STORE", "memory"),
		DynamoTable:    getenv("DDB_TABLE_NAME", "cloudops-main-local"),
		DynamoEndpoint: getenv("AWS_ENDPOINT_URL", ""),
		AWSRegion:      getenv("AWS_REGION", "eu-west-1"),
	}
	if v := os.Getenv("APP_SHUTDOWN_SECONDS"); v != "" {
		n, err := strconv.Atoi(v)
		if err != nil || n <= 0 {
			return Config{}, fmt.Errorf("APP_SHUTDOWN_SECONDS must be a positive integer")
		}
		cfg.ShutdownWait = time.Duration(n) * time.Second
	}
	if v := os.Getenv("APP_SEED_LOCAL"); v != "" {
		cfg.SeedLocalData = v == "1" || strings.EqualFold(v, "true")
	}
	if cfg.Env != "local" {
		cfg.SeedLocalData = getenv("APP_SEED_LOCAL", "") == "true" || os.Getenv("APP_SEED_LOCAL") == "1"
	}
	origins := getenv("APP_CORS_ORIGINS", "http://localhost:4200")
	for _, part := range strings.Split(origins, ",") {
		part = strings.TrimSpace(part)
		if part != "" {
			cfg.CORSOrigins = append(cfg.CORSOrigins, part)
		}
	}
	switch cfg.LogLevel {
	case "debug", "info", "warn", "error":
	default:
		return Config{}, fmt.Errorf("APP_LOG_LEVEL must be debug, info, warn, or error")
	}
	if cfg.HTTPAddr == "" {
		return Config{}, fmt.Errorf("APP_HTTP_ADDR is required")
	}
	switch cfg.StoreDriver {
	case "memory", "dynamodb":
	default:
		return Config{}, fmt.Errorf("APP_STORE must be memory or dynamodb")
	}
	return cfg, nil
}

func getenv(key, fallback string) string {
	if v := strings.TrimSpace(os.Getenv(key)); v != "" {
		return v
	}
	return fallback
}

func loadDotEnv(path string) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		key, val, ok := strings.Cut(line, "=")
		if !ok {
			continue
		}
		key = strings.TrimSpace(key)
		val = strings.Trim(strings.TrimSpace(val), `"'`)
		if os.Getenv(key) == "" {
			_ = os.Setenv(key, val)
		}
	}
	return sc.Err()
}
