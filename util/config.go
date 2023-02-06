package util

import "github.com/spf13/viper"

type Config struct {
	DBSource             string `mapstructure:"DB_SOURCE"`
	RedisSource          string `mapstructure:"REDIS_SOURCE"`
	ServerAddress        string `mapstructure:"SERVER_ADDRESS"`
	TokenKey             string `mapstructure:"TOKEN_KEY"`
	EncryptKey           string `mapstructure:"ENCRYPT_KEY"`
	AccessTokenDuration  string `mapstructure:"ACCESS_TOKEN_DURATION"`
	RefreshTokenDuration string `mapstructure:"REFRESH_TOKEN_DURATION"`
	Salt                 string `mapstructure:"SALT"`
}

func LoadConfig(path string) (config Config, err error) {
	viper.AddConfigPath(path)
	viper.SetConfigName("app")
	viper.SetConfigType("env")

	viper.AutomaticEnv()
	err = viper.ReadInConfig()
	if err != nil {
		return
	}
	err = viper.Unmarshal(&config)
	return
}
