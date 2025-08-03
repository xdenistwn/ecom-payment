package config

import (
	"log"

	"github.com/spf13/viper"
)

func LoadConfig() Config {
	var cfg Config

	viper.AddConfigPath("./files/config")
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	err := viper.ReadInConfig()

	if err != nil {
		log.Fatalf("error read config file: %s", err)
	}

	if err := viper.Unmarshal(&cfg); err != nil {
		log.Fatalf("error unmarshal app config: %s", err)
	}

	return cfg
}
