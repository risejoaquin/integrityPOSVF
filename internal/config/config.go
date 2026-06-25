package config

import (
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Server       ServerConfig       `yaml:"server"`
	Database     DatabaseConfig     `yaml:"database"`
	Printer      PrinterConfig      `yaml:"printer"`
	WhatsApp     WhatsAppConfig     `yaml:"whatsapp"`
	AI           AIConfig           `yaml:"ai"`
	CustomPath   string             `yaml:"custom_path"`
	Modules      map[string]bool    `yaml:"modules"`
	BusinessInfo BusinessInfoConfig `yaml:"business_info"`
}

type BusinessInfoConfig struct {
	Name  string `yaml:"name"`
	Color string `yaml:"color"`
}

type ServerConfig struct {
	Port string `yaml:"port"`
}

type DatabaseConfig struct {
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	User     string `yaml:"user"`
	Password string `yaml:"password"`
	DBName   string `yaml:"dbname"`
	SSLMode  string `yaml:"sslmode"`
}

type PrinterConfig struct {
	Type       string `yaml:"type"`
	DevicePath string `yaml:"device_path"`
	TCPAddr    string `yaml:"tcp_addr"`
}

type WhatsAppConfig struct {
	Enabled       bool   `yaml:"enabled"`
	Token         string `yaml:"token"`
	PhoneNumberID string `yaml:"phone_number_id"`
	VerifyToken   string `yaml:"verify_token"`
}

type AIConfig struct {
	Enabled bool   `yaml:"enabled"`
	APIKey  string `yaml:"api_key"`
	Model   string `yaml:"model"`
}

func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}
