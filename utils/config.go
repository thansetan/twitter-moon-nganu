package utils

import (
	"github.com/spf13/viper"
)

type Config struct {
	ConsumerKey    string `mapstructure:"CONSUMER_KEY"`
	ConsumerSecret string `mapstructure:"CONSUMER_SECRET"`
	CronJobAPIKey  string `mapstructure:"CRONJOB_API_KEY"`
	SessionKey     string `mapstructure:"SESSION_KEY"`
	RedisURL       string `mapstructure:"REDIS_URL"`
}

func LoadConfig(confFile string) (Config, error) {
	var conf Config
	viper := viper.New()

	err := viper.BindEnv(confFile)
	if err != nil {
		return conf, err
	}

	viper.SetConfigFile(confFile)
	viper.AutomaticEnv()

	err = viper.ReadInConfig()
	if err != nil {
		return conf, err
	}

	err = viper.Unmarshal(&conf)
	if err != nil {
		return conf, err
	}

	return conf, nil
}
