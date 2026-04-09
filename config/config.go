package config

import (
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	Port string
	Env  string

	DBHost     string
	DBPort     string
	DBUser     string
	DBPassword string
	DBName     string

	JWTSecret         string
	JWTAccessExpires  time.Duration
	JWTRefreshExpires time.Duration

	FrontEndUrl    string
	ReturnUrl      string
	SteamOpenIdUrl string
	SteamApiKey    string
	SteamApiUrl    string
}

var App *Config

func Load() {
	cwd, err := os.Getwd()
	if err != nil {
		log.Printf("Config Load: unable to get current working directory: %v", err)
	}

	_, b, _, _ := runtime.Caller(0)
	basepath := filepath.Join(filepath.Dir(b), "..")

	envPaths := []string{
		filepath.Join(cwd, ".env"),
		filepath.Join(basepath, ".env"),
	}

	loaded := false
	for _, path := range envPaths {
		if path == "" {
			continue
		}

		if err := godotenv.Load(path); err != nil {
			if os.IsNotExist(err) || strings.Contains(err.Error(), "no such file") {
				continue
			}
			log.Printf("Config Load: failed to load .env from %s: %v", path, err)
			break
		}
		loaded = true
		break
	}

	if !loaded {
		log.Printf("Config Load: no .env file loaded, checked paths: %v", envPaths)
	}

	accessExpires, err := time.ParseDuration(getEnv("JWT_ACCESS_EXPIRES", "15m"))
	if err != nil {
		accessExpires = 15 * time.Minute
	}

	refreshExpires, err := time.ParseDuration(getEnv("JWT_REFRESH_EXPIRES", "720h"))
	if err != nil {
		refreshExpires = 30 * 24 * time.Hour
	}

	App = &Config{
		Port: getEnv("PORT", "8080"),
		Env:  getEnv("ENV", "development"),

		DBHost:     getEnv("DB_HOST", "localhost"),
		DBPort:     getEnv("DB_PORT", "3306"),
		DBUser:     getEnv("DB_USER", "root"),
		DBPassword: getEnv("DB_PASSWORD", ""),
		DBName:     getEnv("DB_NAME", "gamelog"),

		JWTSecret:         getEnv("JWT_SECRET", "fallback-secret"),
		JWTAccessExpires:  accessExpires,
		JWTRefreshExpires: refreshExpires,
		FrontEndUrl:       getEnv("FRONTEND_URL", "http://localhost:4200"),
		ReturnUrl:         getEnv("RETURN_URL", "http://localhost:8080/api/v1/auth/steam/callback"),
		SteamOpenIdUrl:    getEnv("STEAM_OPEN_ID_URL", "https://steamcommunity.com/openid/login"),
		SteamApiKey:       getEnv("STEAM_API_KEY", ""),
		SteamApiUrl:       getEnv("STEAM_API_URL", "https://api.steampowered.com/ISteamUser/GetPlayerSummaries/v0002/?key=%s&steamids=%s"),
	}
}

func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
