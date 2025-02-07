package common

import (
	"fmt"

	"github.com/spf13/viper"
)

type config struct {
	LogFile    string     `mapstructure:"log.file"`
	DBName     string     `mapstructure:"db_name"`
	Domain     string     `mapstructure:"domain"`
	Debug      bool       `mapstructure:"debug"`
	SMTPServer string     `mapstructure:"smtp.server"`
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
	viper.SetDefault("log.file", "castellers.log")
	viper.SetDefault("log.level", "info")
	viper.SetDefault("db_name", "castellers.db")
	viper.SetDefault("domain", "localhost")
	viper.SetDefault("debug", false)
	viper.SetDefault("smtp.server", "127.0.0.1")
	viper.SetDefault("smtp.port", "25")
	viper.SetDefault("smtp.username", "")
	viper.SetDefault("smtp.password", "")
	viper.SetDefault("smtp.enabled", true)
	viper.SetDefault("reminder_time_before_event", 172800)   // 2 days
	viper.SetDefault("summary_time_before_event", 86400)     // 1 day
	viper.SetDefault("encryption.iterations", 10000)         // For hashing encryption key
	viper.SetDefault("encryption.password_hashing_cost", 10) // For hashing passwords
	viper.SetDefault("redis_dsn", "localhost:6379")          // Redis connection
	viper.SetDefault("jwt.access_ttl_minutes", 15)
	viper.SetDefault("jwt.refresh_ttl_days", 15)
	viper.SetDefault("jwt.reset_ttl_minutes", 60)
	viper.SetDefault("jwt.participation_ttl_minutes", 2880)
	viper.SetDefault("jwt.registration_ttl_minutes", 10080)
	viper.SetDefault("inactive_delay_days", 21)
	viper.SetDefault("otel_enable", false)

	// read config file
	err := viper.ReadInConfig()
	if err != nil {
		fmt.Println("WARN: cannot read config file")
	}

	// read environment variables
	viper.SetEnvPrefix("app")
	viper.BindEnv("log.file")
	viper.BindEnv("log.level", "APP_LOG_LEVEL")
	viper.BindEnv("db_name")
	viper.BindEnv("domain")
	viper.BindEnv("cdn")
	viper.BindEnv("debug")
	viper.BindEnv("smtp.server", "APP_SMTP_SERVER")
	viper.BindEnv("smtp.port", "APP_SMTP_PORT")
	viper.BindEnv("smtp.username", "APP_SMTP_USERNAME")
	viper.BindEnv("smtp.password", "APP_SMTP_PASSWORD")
	viper.BindEnv("smtp.enabled", "APP_SMTP_ENABLED")
	viper.BindEnv("reply_to", "APP_REPLY_TO")
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
	viper.BindEnv("jwt.reset_ttl_minutes", "APP_RESET_TTL_MINUTES")
	viper.BindEnv("jwt.participation_ttl_minutes", "APP_PARTICIPATION_TTL_MINUTES")
	viper.BindEnv("jwt.registration_ttl_minutes", "APP_REGISTRATION_TTL_MINUTES")
	viper.BindEnv("otel_enable", "APP_OTEL_ENABLE")
	viper.BindEnv("inactive_delay_days", "APP_INACTIVE_DELAY_DAYS")

	var c config
	err = viper.Unmarshal(&c)
	if err != nil {
		panic(fmt.Errorf("unable to parse configuration, %v", err))
	}

	if !viper.IsSet("cdn") {
		viper.Set("cdn", viper.GetString("domain"))
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
