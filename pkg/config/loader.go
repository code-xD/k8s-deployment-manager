package config

import (
	"os"

	"github.com/code-xd/k8s-deployment-manager/pkg/constants"
	"github.com/spf13/viper"
)

type ConfigLoader[T any] struct {
	fileName string
	filePath string
}

func (c *ConfigLoader[T]) getEnv() string {
	env := os.Getenv(constants.APP_ENV_VAR)
	if env == "" {
		return constants.APP_ENV_DEV
	}

	return env
}

// get config file path from arg
func (c *ConfigLoader[T]) Load() (*T, error) {

	env := c.getEnv()
	viper.SetConfigName(c.fileName + "." + env)
	viper.SetConfigType("yaml")
	viper.AddConfigPath(c.filePath)

	if err := viper.ReadInConfig(); err != nil {
		return nil, err
	}

	var cfg T
	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}

func NewConfigLoader[T any](fileName string, filePath string) *ConfigLoader[T] {
	return &ConfigLoader[T]{
		fileName: fileName,
		filePath: filePath,
	}
}
