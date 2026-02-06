package main

import (
	"github.com/begenov/real-estate/internal/app"
	"github.com/begenov/real-estate/internal/config"
	"github.com/begenov/real-estate/internal/logger"
)

const (
	configsDir = "config"
	configName = "app"
	configType = "env"
)

func main() {

	cfg, err := config.NewConfig(configsDir, configName, configType)
	if err != nil {
		logger.Error("config.NewConfig(): ", err)
		return
	}

	if err := app.Run(cfg); err != nil {
		logger.Error("app.Run(): ", err)
	}
}
