package main

import (
	"github.com/joho/godotenv"
	"os"
	"strconv"
	"strings"
)

type Uptimerobot struct {
	Token []string
	Email []string
}

// Environment is a main environment structure
type Environment struct {
	Uptimerobot Uptimerobot
}

// New returns a new Config struct
func NewEnv() *Environment {
	return &Environment{
		Uptimerobot: Uptimerobot{
			Token: getEnvAsSlice("UPTIMEROBOT_TOKEN", []string{}, ","),
			Email: getEnvAsSlice("UPTIMEROBOT_EMAIL", []string{}, ","),
		},
	}
}

// loadEnv trying to load .env file, or write error to log
func loadEnv() {
	_ = godotenv.Load()
}

// getEnvAsString receives variable name that we need
// from environment and returns its value from .env file
// or defaultValue if function can't find this variable name in .env file
func getEnvAsString(neededValue string, defaultValue string) string {
	loadEnv()
	if value, exists := os.LookupEnv(neededValue); exists {
		return value
	}
	return defaultValue
}

// getEnvAsInt receives variable name that we need
// from environment and returns its value as int from .env file
// or defaultValue if function can't find this variable name in .env file
func getEnvAsInt(neededValue string, defaultValue int) int {
	valueInString := getEnvAsString(neededValue, "")
	if value, err := strconv.Atoi(valueInString); err == nil {
		return value
	}
	return defaultValue
}

// getEnvAsBool receives variable name that we need
// from environment and returns its value as bool from .env file
// or defaultValue if function can't find this variable name in .env file
func getEnvAsBool(neededValue string, defaultValue bool) bool {
	valueInString := getEnvAsString(neededValue, "")
	if value, err := strconv.ParseBool(valueInString); err == nil {
		return value
	}
	return defaultValue
}

// getEnvAsSlice receives variable name that we need
// from environment and returns its value as slice from .env file
// or defaultValue if function can't find this variable name in .env file
func getEnvAsSlice(neededValue string, defaultValue []string, separator string) []string {
	valueInString := getEnvAsString(neededValue, "")
	if valueInString == "" {
		return defaultValue
	}
	value := strings.Split(valueInString, separator)
	return value
}
