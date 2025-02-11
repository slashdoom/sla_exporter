package main

import (
	"bytes"
	"io/ioutil"

	"example.org/config"
	"example.org/logger"

	"go.uber.org/zap"
	"gopkg.in/yaml.v3"
)

func MergeConfig(c config.Config) (config.Config, error) {
	f, err := loadConfigFromFile(c.ConfigFile)
	if err != nil {
		logger.Log.Warn("config file error, exiting",
			zap.String("file", c.ConfigFile),
			zap.Error(err),
		)
		return c, err
	}

	logger.Log.Info("loading config flags")
	c = loadConfigFromFlags(c, f)

	return c, nil
}

func loadConfigFromFile(f string) (config.Config, error) {
	c := config.New()
	logger.Log.Info("loading config from file",
		zap.String("file", f),
	)

	a, err := ioutil.ReadFile(f)
	if err != nil {
		logger.Log.Warn("error reading config file",
			zap.String("file", f),
			zap.Error(err),
		)
		return c, err
	}

	b, err := ioutil.ReadAll(bytes.NewReader(a))
	if err != nil {
		logger.Log.Warn("error loading config file",
			zap.String("file", f),
			zap.Error(err),
		)
		return c, err
	}

	err = yaml.Unmarshal(b, &c)
	if err != nil {
		logger.Log.Warn("error parsing config file",
			zap.String("file", f),
			zap.Error(err),
		)
		return c, err
	}

	c.ConfigFile = f

	return c, nil
}

func loadConfigFromFlags(c config.Config, f config.Config) config.Config {
	if c.Web.ListenAddress < 1 {
		f.Web.ListenAddress = c.Web.ListenAddress
	}
	if c.Web.MetricsPath != "" {
		f.Web.MetricsPath = c.Web.MetricsPath
	}
	if c.Timeout != "" {
		f.Timeout = c.Timeout
	}
	if c.Level != "" {
		f.Level = c.Level
	}

	return f
}
