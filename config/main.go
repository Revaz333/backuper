package config

import (
	"fmt"

	"github.com/spf13/viper"
)

type (
	Config struct {
		Services map[string]ConfigService
		Drivers  struct {
			S3 map[string]struct {
				Region     string
				Access_Key string
				Secret_Key string
				Endpoint   string
				Bucket     string
			}

			FTP map[string]struct {
				Host     string
				User     string
				Passwrod string
				Port     int
				Folder   string
			}
		}
	}

	ConfigService struct {
		Target_Folder string
		File_Name     string
		Archiver      string // default tar
		Driver        struct {
			Type    string
			Profile string
		}
		Excluded_Dirs []string
		Frequency     int64
		Spec          string
		Max_Images    int
	}
)

func New() *Config {

	return &Config{}
}

func (cfg *Config) LoadConfig() error {

	viper.SetConfigFile("/etc/backuper/config.yml")

	cfg.setDefaults()

	if err := viper.ReadInConfig(); err != nil {
		return fmt.Errorf("failed to read config: %v", err)
	}

	if err := viper.Unmarshal(&cfg); err != nil {
		return fmt.Errorf("failed to unpack config: %v", err)
	}

	return nil
}
