package config

import (
	"os"
	"strconv"
)

// Config holds the application configuration
type Config struct {
	Server   ServerConfig
	Database DatabaseConfig
	VNPay    VNPayConfig
}

// ServerConfig holds the server configuration
type ServerConfig struct {
	Port string
}

// DatabaseConfig holds the database configuration
type DatabaseConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	DBName   string
	SSLMode  string
}

// VNPayConfig holds the configuration for VNPAY integration
type VNPayConfig struct {
	TmnCode        string
	HashSecret     string
	VNPayURL       string
	ReturnURL      string
	APIUrl         string
	MerchantAPI    string
	TransactionAPI string
}

// LoadConfig loads configuration from environment variables
func LoadConfig() *Config {
	return &Config{
		Server: ServerConfig{
			Port: getEnv("SERVER_PORT", "8080"),
		},
		Database: DatabaseConfig{
			Host:     getEnv("DB_HOST", "localhost"),
			Port:     getEnv("DB_PORT", "5432"),
			User:     getEnv("DB_USER", "postgres"),
			Password: getEnv("DB_PASSWORD", "postgres"),
			DBName:   getEnv("DB_NAME", "payment_service"),
			SSLMode:  getEnv("DB_SSLMODE", "disable"),
		},
		VNPay: VNPayConfig{
			TmnCode:        getEnv("VNPAY_TMN_CODE", ""),
			HashSecret:     getEnv("VNPAY_HASH_SECRET", ""),
			VNPayURL:       getEnv("VNPAY_URL", "https://sandbox.vnpayment.vn/paymentv2/vpcpay.html"),
			ReturnURL:      getEnv("VNPAY_RETURN_URL", "http://localhost:8080/api/vnpay/return"),
			APIUrl:         getEnv("VNPAY_API_URL", "http://sandbox.vnpayment.vn/merchant_webapi/merchant.html"),
			TransactionAPI: getEnv("VNPAY_TRANSACTION_API", "https://sandbox.vnpayment.vn/merchant_webapi/api/transaction"),
		},
	}
}

// GetDatabaseDSN returns the database connection string
func (c *DatabaseConfig) GetDatabaseDSN() string {
	return "postgres://" + c.User + ":" + c.Password + "@" + c.Host + ":" + c.Port + "/" + c.DBName + "?sslmode=" + c.SSLMode
}

// Helper function to get environment variable with a default value
func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}

// Helper function to get integer environment variable with a default value
func getEnvAsInt(key string, defaultValue int) int {
	valueStr := getEnv(key, "")
	if value, err := strconv.Atoi(valueStr); err == nil {
		return value
	}
	return defaultValue
}
