package config

import (
	"fmt"
)

// DBConfig holds database configuration details
type DBConfig struct {
	Host     string
	Port     int
	User     string
	Password string
	DBName   string
	DBSchema string
	SSLMode  string
}

func (config DBConfig) Print() {
	fmt.Println("Database Configuration:")
	fmt.Println("-----------------------")
	fmt.Printf("Host: %s\n", config.Host)
	fmt.Printf("Port: %d\n", config.Port)
	fmt.Printf("User: %s\n", config.User)
	fmt.Printf("Password: %s\n", config.Password) // Be cautious about logging passwords
	fmt.Printf("Database Name: %s\n", config.DBName)
	fmt.Printf("Database Schema: %s\n", config.DBSchema)
	fmt.Printf("SSL mode: %s\n", config.SSLMode)
	fmt.Println("-----------------------")
}
