package config

import (
	"fmt"
	"strings"

	"github.com/spf13/viper"
)

type Config struct {
	AppPort                string `mapstructure:"APP_PORT"`
	AppEnv                 string `mapstructure:"APP_ENV"`
	MongoURI               string `mapstructure:"MONGODB_URI"`
	MongoDB                string `mapstructure:"MONGODB_DATABASE"`
	ProductsSourceEndpoint string `mapstructure:"PRODUCTS_SOURCE_ENDPOINT"`
	SwaggerBaseURL         string `mapstructure:"SWAGGER_BASE_URL"`
	SwaggerScheme          string `mapstructure:"SWAGGER_SCHEME"`
}

func Load() (*Config, error) {
	viper.SetConfigFile(".env")
	viper.SetConfigType("env")
	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	_ = viper.ReadInConfig()

	viper.SetDefault("APP_PORT", "8080")
	viper.SetDefault("APP_ENV", "development")
	viper.SetDefault("MONGODB_URI", "mongodb://localhost:27017")
	viper.SetDefault("MONGODB_DATABASE", "foodji")
	viper.SetDefault("SWAGGER_BASE_URL", "localhost:8080")
	viper.SetDefault("SWAGGER_SCHEME", "http")

	cfg := &Config{}
	if err := viper.Unmarshal(cfg); err != nil {
		return nil, fmt.Errorf("unmarshal config: %w", err)
	}

	return cfg, nil
}
