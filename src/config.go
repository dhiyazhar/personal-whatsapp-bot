package main

import (
	"fmt"
	"os"

	_ "github.com/joho/godotenv/autoload"
)

type DatabaseConfig struct {
	User     string
	Password string
	Host     string
	Port     string
	DBName   string
}

func (d *DatabaseConfig) GetConnectionString() string {
	return fmt.Sprintf("postgresql://%s:%s@%s:%s/%s?sslmode=require",
		d.User, d.Password, d.Host, d.Port, d.DBName)
}

func LoadConfig() (*DatabaseConfig, error) {
	config := &DatabaseConfig{
		User:     os.Getenv("SUPABASE_USER"),
		Password: os.Getenv("SUPABASE_PASSWORD"),
		Host:     os.Getenv("SUPABASE_HOST"),
		Port:     os.Getenv("SUPABASE_PORT"),
		DBName:   os.Getenv("SUPABASE_DBNAME"),
	}

	return config, nil
}
