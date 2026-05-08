package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

const (
	defaultAppName     = "report-service"
	defaultEnvironment = "local"
	defaultPort        = "8080"
	defaultBodyLimitMB = 10
)

type Config struct {
	AppName     string
	Environment string
	Port        string
	BodyLimit   int
}

func Load() Config {
	bodyLimitMB := envInt("BODY_LIMIT_MB", defaultBodyLimitMB)
	if bodyLimitMB <= 0 {
		bodyLimitMB = defaultBodyLimitMB
	}

	return Config{
		AppName:     envString("APP_NAME", defaultAppName),
		Environment: envString("APP_ENV", defaultEnvironment),
		Port:        normalizePort(envString("PORT", defaultPort)),
		BodyLimit:   bodyLimitMB << 20,
	}
}

func (c Config) Addr() string {
	return fmt.Sprintf(":%s", normalizePort(c.Port))
}

func envString(key, fallback string) string {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}
	return value
}

func envInt(key string, fallback int) int {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}

	number, err := strconv.Atoi(value)
	if err != nil {
		return fallback
	}

	return number
}

func normalizePort(value string) string {
	value = strings.TrimSpace(value)
	value = strings.TrimPrefix(value, ":")
	if value == "" {
		return defaultPort
	}
	return value
}
