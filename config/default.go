package config

import "github.com/spf13/viper"

func (cfg *Config) setDefaults() {

	// Here we set config variables default values

	viper.SetDefault("APP.PORT", "8080")
	viper.SetDefault("APP.DEBUG", true)
	viper.SetDefault("APP.DBCONN", "")
	viper.SetDefault("APP.CODEEXPIRETIME", "")
	viper.SetDefault("APP.CODENEXTSENDTIME", "")
	viper.SetDefault("SMS.APIURL", "")
	viper.SetDefault("SMS.TPLPATH", "resources/templates/sms")
	// viper.SetDefault("PUSHTPLPATH", "resources/templates/push")
	viper.SetDefault("JWTSECRET", "")
	viper.SetDefault("SMSEXP", "")
	viper.SetDefault("TELEGRAM.TOKEN", "")
	viper.SetDefault("TELEGRAM.CHATID", -0)
}
