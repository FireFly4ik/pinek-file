package config

import (
	"fmt"
	"os"
	"strings"
)

type Config struct {
	Address        string
	Port           string
	ProductionType string

	Consul Consul
	Minio  Minio
}

type Consul struct {
	Address       string
	Name          string
	CheckId       string
	RegisterTTL   string
	RefreshTTL    string
	DeregisterTTL string
}

type Minio struct {
	IP           string
	Port         string
	ConsolePort  string
	RootUser     string
	RootPassword string
	BucketName   string
	UseSSL       string
}

func NewEnvConfig() *Config {
	return &Config{
		Address:        os.Getenv("ADDRESS"),
		Port:           os.Getenv("PORT"),
		ProductionType: os.Getenv("PRODUCTION_TYPE"),

		Consul: Consul{
			Address:       os.Getenv("CONSUL_ADDRESS"),
			Name:          os.Getenv("CONSUL_SERVICE_NAME"),
			CheckId:       os.Getenv("CONSUL_CHECK_ID"),
			RegisterTTL:   os.Getenv("CONSUL_REGISTER_TTL"),
			RefreshTTL:    os.Getenv("CONSUL_REFRESH_TTL"),
			DeregisterTTL: os.Getenv("CONSUL_DEREGISTER_TTL"),
		},

		Minio: Minio{
			IP:           os.Getenv("MINIO_IP"),
			Port:         os.Getenv("MINIO_PORT"),
			ConsolePort:  os.Getenv("MINIO_CONSOLE_PORT"),
			RootUser:     os.Getenv("MINIO_ROOT_USER"),
			RootPassword: os.Getenv("MINIO_ROOT_PASSWORD"),
			BucketName:   os.Getenv("MINIO_BUCKET_NAME"),
			UseSSL:       os.Getenv("MINIO_USE_SSL"),
		},
	}
}

func PrintConfigWithHiddenSecrets(config *Config) {
	// Функция для маскировки секретов
	mask := func(s string) string {
		if s == "" {
			return ""
		}
		return strings.Repeat("*", len(s))
	}

	fmt.Println("========== Configuration ==========\n")

	fmt.Println("App Configuration:")
	fmt.Printf("\tAddress: %s\n", config.Address)
	fmt.Printf("\tPort: %s\n", config.Port)
	fmt.Printf("\tProductionType: %s\n", config.ProductionType)

	fmt.Println("\nConsul Configuration:")
	fmt.Printf("\tAddress: %s\n", config.Consul.Address)
	fmt.Printf("\tName: %s\n", config.Consul.Name)
	fmt.Printf("\tCheckId: %s\n", config.Consul.CheckId)
	fmt.Printf("\tRegisterTTL: %s\n", config.Consul.RegisterTTL)
	fmt.Printf("\tRefreshTTL: %s\n", config.Consul.RefreshTTL)
	fmt.Printf("\tDeregisterTTL: %s\n", config.Consul.DeregisterTTL)

	fmt.Println("\nMinio Configuration:")
	fmt.Printf("\tIP: %s\n", config.Minio.IP)
	fmt.Printf("\tPort: %s\n", config.Minio.Port)
	fmt.Printf("\tConsolePort: %s\n", config.Minio.ConsolePort)
	fmt.Printf("\tRootUser: %s\n", config.Minio.RootUser)
	fmt.Printf("\tRootPassword: %s\n", mask(config.Minio.RootPassword))
	fmt.Printf("\tBucketName: %s\n", config.Minio.BucketName)
	fmt.Printf("\tUseSSL: %s\n", config.Minio.UseSSL)

	fmt.Println("\n===================================")
}
