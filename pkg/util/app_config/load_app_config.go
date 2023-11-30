package app_config

import (
	"strings"

	"github.com/frontierdigital/sheriff/pkg/core"
	"github.com/spf13/viper"
)

func LoadAppConfig() (appConfig *core.AppConfig, err error) {
	viper.SetEnvPrefix("SHERIFF")
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.AutomaticEnv()
	err = viper.Unmarshal(&appConfig)

	return
}
