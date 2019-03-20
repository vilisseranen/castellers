package common

import (
	"fmt"
	"github.com/spf13/viper"
)

type config struct {
	LogFile    string `mapstructure:"log_file"`
	DBName     string `mapstructure:"db_name"`
	Domain     string `mapstructure:"domain"`
	Debug      bool   `mapstructure:"debug"`
	SMTPServer string `mapstructure:"smtp_server"`
}

func ReadConfig() config {

	// config file location
	viper.SetConfigName("config")
	viper.AddConfigPath("/etc/castellers/")
	viper.AddConfigPath(".")

	// setting defaults
	viper.SetDefault("log_file", "castellers.log")
	viper.SetDefault("db_name", "castellers.db")
	viper.SetDefault("domain", "localhost")
	viper.SetDefault("debug", false)
	viper.SetDefault("smtp_server", "127.0.0.1:25")
	viper.SetDefault("mail_from", "clement@clemissa.info")
	viper.SetDefault("notification_time_before_event", 172800) // 2 days

	// read config file
	err := viper.ReadInConfig()
	if err != nil {
		fmt.Println("WARN: cannot read config file")
	}

	// read environment variables
	viper.SetEnvPrefix("app")
	viper.BindEnv("log_file")
	viper.BindEnv("db_name")
	viper.BindEnv("domain")
	viper.BindEnv("debug")
	viper.BindEnv("smtp_server")
	viper.BindEnv("mail_from")
	viper.BindEnv("notification_time_before_event")

	var c config
	err = viper.Unmarshal(&c)
	if err != nil {
		panic(fmt.Errorf("Unable to parse configuration, %v\n", err))
	}

	if c.Debug {
		fmt.Println("Config parsed:")
		fmt.Printf("%+v", c)
		fmt.Println()
	}
	return c
}

func GetConfigString(key string) string {
	return viper.GetString(key)
}

func GetConfigBool(key string) bool {
	return viper.GetBool(key)
}

func GetConfigInt(key string) int {
	return viper.GetInt(key)
}
