package common

import (
	"fmt"

	"github.com/spf13/viper"
)

type config struct {
	LogFile    string     `mapstructure:"log_file"`
	DBName     string     `mapstructure:"db_name"`
	Domain     string     `mapstructure:"domain"`
	Debug      bool       `mapstructure:"debug"`
	SMTPServer string     `mapstructure:"smtp_server"`
	Encryption encryption `mapstructure:"encryption"`
}

type encryption struct {
	Key            string `mapstructure:"key"`
	KeySalt        string `mapstructure:"key_salt"`
	PasswordPepper string `mapstructure:"password_pepper"`
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
	viper.SetDefault("reminder_time_before_event", 172800)   // 2 days
	viper.SetDefault("summary_time_before_event", 86400)     // 1 day
	viper.SetDefault("encryption.iterations", 10000)         // For hashing encryption key
	viper.SetDefault("encryption.password_hashing_cost", 10) // For hashing passwords
	viper.SetDefault("redis_dsn", "localhost:6379")          // Redis connection
	viper.SetDefault("jwt.access_ttl_minutes", 15)
	viper.SetDefault("jwt.refresh_ttl_days", 15)

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
	viper.BindEnv("redis_dsn")
	viper.BindEnv("reminder_time_before_event")
	viper.BindEnv("summary_time_before_event")
	viper.BindEnv("encryption.key", "APP_KEY")
	viper.BindEnv("encryption.key_salt", "APP_KEY_SALT")
	viper.BindEnv("encryption.password_pepper", "APP_PASSWORD_PEPPER")
	viper.BindEnv("jwt.access_secret", "APP_ACCESS_SECRET")
	viper.BindEnv("jwt.refresh_secret", "APP_REFRESH_SECRET")
	viper.BindEnv("jwt.access_ttl_minutes", "APP_ACCESS_TTL_MINUTES")
	viper.BindEnv("jwt.refresh_ttl_days", "APP_REFRESH_TTL_DAYS")

	var c config
	err = viper.Unmarshal(&c)
	if err != nil {
		panic(fmt.Errorf("Unable to parse configuration, %v", err))
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
