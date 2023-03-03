package config

import (
	"fmt"
	"github.com/pkg/errors"
	"github.com/spf13/viper"
	"os"
	"strings"
)

var ServiceConf = &ServiceConfig{}

const (
	Local      = "local"
	Dev        = "dev"
	Qa         = "qa"
	Uat        = "uat"
	Production = "prod"
	Default    = Dev
)

type ServiceConfig struct {
	AppConfig           AppConfig           `yaml:"appConfig"`
	DBConfig            DBConfig            `yaml:"dBConfig"`
	HTTPClientConfig    HTTPClientConfig    `yaml:"httpClientConfig"`
	OpenTelemetryConfig OpenTelemetryConfig `yaml:"openTelemetryConfig"`
	KafkaConfig         KafkaConfig         `yaml:"kafkaConfig"`
}

type KafkaConfig struct {
	Brokers       []string `yaml:"brokers"`
	ConsumerGroup string   `yaml:"consumerGroup"`
	MaxRetries    int      `yaml:"maxRetries"`
}

type OpenTelemetryConfig struct {
	CollectorEndpoint string `yaml:"collectorEndpoint"`
}

type DBConfig struct {
	LocationsDatabaseConnection string `yaml:"locationsDatabaseConnection"`
	ConnMaxIdleTime             int    `yaml:"connMaxIdleTime"`
	MaxIdleConns                int    `yaml:"maxIdleConns"`
	ConnMaxLifetime             int    `yaml:"connMaxLifetime"`
}

type HTTPClientConfig struct {
	MaxIdleConns           int `yaml:"maxIdleConns"`
	MaxConnsPerHost        int `yaml:"maxConnsPerHost"`
	MaxIdleConnsPerHost    int `yaml:"maxIdleConnsPerHost"`
	IdleConnTimeoutSeconds int `yaml:"idleConnTimeoutSeconds"`
	RequestTimeoutSeconds  int `yaml:"requestTimeoutSeconds"`
}

type AppConfig struct {
	Name        string `yaml:"name"`
	Version     string `yaml:"version"`
	BindAddress string `yaml:"bindAddress"`
}

func LoadConfig() (*ServiceConfig, error) {
	// Get the environment to run in...
	env, _ := GetEnvironment()
	configFile := fmt.Sprintf("service-%s", env)

	viperInstance := viper.New()
	viperInstance.AutomaticEnv()
	viperInstance.AddConfigPath("config")
	viperInstance.AddConfigPath("../config")
	viperInstance.AddConfigPath("../../config")
	viperInstance.SetConfigName(configFile)

	// Try to load the config file for the environment
	if err := viperInstance.ReadInConfig(); err != nil {
		return nil, errors.Wrapf(err, "Failed to read configuration file")
	}

	ServiceConf = &ServiceConfig{}
	if err := viperInstance.Unmarshal(ServiceConf); err != nil {
		return nil, errors.Wrapf(err, "Failed to parse configuration")
	}

	return ServiceConf, nil
}

func GetCorsOriginAddressByEnv(env string) []string {
	var allowedOrigins []string
	switch env {
	case Dev:
		allowedOrigins = []string{"*"}
	case Qa:
		allowedOrigins = []string{"*"}
	case Uat:
		allowedOrigins = []string{"*"}
	case Production:
		allowedOrigins = []string{"*"}
	default:
		allowedOrigins = []string{"*"}
	}

	return allowedOrigins
}

func GetIntValueOrDefault(value, defaultValue int) int {
	if value == 0 {
		return defaultValue
	}
	return value
}

func GetEnvironment() (string, error) {
	if envName := os.Getenv("ENV"); len(envName) > 0 {
		env := strings.ToLower(envName)
		switch env {
		case Dev, Qa, Local, Production, Uat:
			return env, nil
		default:
			return Default, errors.New("env not defined")
		}
	}

	return Default, nil
}
