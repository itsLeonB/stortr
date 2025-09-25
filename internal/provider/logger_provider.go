package provider

import (
	"github.com/itsLeonB/ezutil/v2"
)

func ProvideLogger(appName string, env string) ezutil.Logger {
	minLevel := 0
	if env == "prod" {
		minLevel = 1
	}

	return ezutil.NewSimpleLogger(appName, true, minLevel)
}
