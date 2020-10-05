package config

import (
	"log"
	"os"
	"strings"

	"github.com/spf13/viper"
)

var defaultConfig *viper.Viper

func ReadConfig(appName string, cfgFile string) *viper.Viper {
	v := viper.New()
	// Defines Prefix for ENV variables and parses them by default
	v.SetEnvPrefix(appName)

	v.AutomaticEnv()
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	// global defaults (key value) - need trailing '/'
	v.SetDefault("defaults.userpath", "/home/")
	// v.SetDefault("loglevel", "debug")

	if cfgFile != "" {
		// Use config file from the flag.
		v.SetConfigFile(cfgFile)
	} else {
		// Find home directory.
		home, err := os.UserHomeDir()
		if err != nil {
			log.Printf("%s", err)
		}

		// Search config in home directory with name ".sftppush" (without extension).
		v.AddConfigPath(home)
		v.SetConfigName(".sftppush")
	}

	// Read config file and not found action
	if err := v.ReadInConfig(); err == nil {
		log.Printf("INFO[+] Using config file: %s\n", v.ConfigFileUsed())

	} else {
		log.Printf("ERROR[-] %s", err)
	}

	// Use if existing config exists
	// v.MergeConfigMap(cfg map[string]interface{})

	return v
}

// TODO config file Validation
// from Hugo https://github.com/gohugoio/hugo/blob/master/config/configLoader.go
// var (
// 	ValidConfigFileExtensions                    = []string{"toml", "yaml", "yml", "json"}
// 	validConfigFileExtensionsMap map[string]bool = make(map[string]bool)
// )

// func init() {
// 	for _, ext := range ValidConfigFileExtensions {
// 		validConfigFileExtensionsMap[ext] = true
// 	}
// }

// // IsValidConfigFilename returns whether filename is one of the supported
// // config formats in Hugo.
// func IsValidConfigFilename(filename string) bool {
// 	ext := strings.ToLower(strings.TrimPrefix(filepath.Ext(filename), "."))
// 	return validConfigFileExtensionsMap[ext]
// }

// // FromConfigString creates a config from the given YAML, JSON or TOML config. This is useful in tests.
// func FromConfigString(config, configType string) (Provider, error) {
// 	v := newViper()
// 	m, err := readConfig(metadecoders.FormatFromString(configType), []byte(config))
// 	if err != nil {
// 		return nil, err
// 	}

// 	v.MergeConfigMap(m)

// 	return v, nil
// }
