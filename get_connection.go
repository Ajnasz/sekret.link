package main

import "os"

func getConnectionString(defaultValue string, envName string) string {
	envURL := os.Getenv(envName)
	if envURL != "" {
		return envURL
	}

	return defaultValue
}
