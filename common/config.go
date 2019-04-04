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

func ReadConfig() {
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
	viper.SetDefault("reminder_time_before_event", 172800) // 2 days
	viper.SetDefault("summary_time_before_event", 86400)   // 1 day

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
	viper.BindEnv("reminder_time_before_event")
	viper.BindEnv("summary_time_before_event")

	var c config
	err = viper.Unmarshal(&c)
	if err != nil {
		panic(fmt.Errorf("Unable to parse configuration, %v", err))
	}

	if c.Debug {
		fmt.Println("Config parsed:")
		fmt.Printf("%+v", c)
		fmt.Println()
	}
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
