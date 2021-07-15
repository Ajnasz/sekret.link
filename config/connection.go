package config

import "os"

func getFromEnvOrDefault(envName string, defaultValue string) string {
	envURL := os.Getenv(envName)
	if envURL != "" {
		return envURL
	}

	return defaultValue
}

// GetConnectionString Returns value from env or the default value
func GetConnectionString(defaultValue string) string {
	return getFromEnvOrDefault("POSTGRES_URL", defaultValue)
}
