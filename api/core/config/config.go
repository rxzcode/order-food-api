package config

import (
	"log"

	"gopkg.in/ini.v1"
)

type AppConfig struct {
	Port string
}

type DBConfig struct {
	User     string
	Password string
	Host     string
	Port     string
	Name     string
}

type AuthConfig struct {
	ApiKey string
}

type Config struct {
	App      AppConfig
	Database DBConfig
	Auth     AuthConfig
}

var Cfg *Config

func LoadConfig(path string) *Config {
	if Cfg != nil {
		return Cfg
	}

	cfg := new(Config)
	iniFile, err := ini.Load(path)
	if err != nil {
		log.Fatalf("Fail to read config file: %v", err)
	}

	err = iniFile.MapTo(cfg)
	if err != nil {
		log.Fatalf("Fail to map config: %v", err)
	}

	Cfg = cfg
	return cfg
}

func GetConfig() *Config {
	if Cfg == nil {
		log.Fatal("Config is not initialized. Call LoadConfig(path) once before using GetConfig().")
	}
	return Cfg
}
