package logger

import (
	"fmt"

	"example.org/config"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var Log *zap.Logger

func init() {
	// Define the log level.
    logLevel := zap.NewAtomicLevel()
	
	// Configure logger
	config := zap.NewDevelopmentConfig()
	config.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	config.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	config.Level = logLevel
	config.OutputPaths = []string{"stdout"} // Send logs to stdout
	var err error
	Log, err = config.Build()
	if err != nil {
		fmt.Println("Error setting up logging:", err)
		return
	}
}

func Config() {
	// Define the log level.
	var err error
    logLevel := zap.NewAtomicLevel()

    // Set the default log level to 'info'.
    level, err := zap.ParseAtomicLevel(config.AppConfig.Level)
	if err != nil {
		Log.Warn("error reconfiguring logging",
			zap.Error(err),
		)
		return
	}
	logLevel = level
	
	// Configure logger
	config := zap.NewDevelopmentConfig()
	config.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	config.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	config.Level = logLevel
	config.OutputPaths = []string{"stdout"} // Send logs to stdout
	
	Log, err = config.Build()
	if err != nil {
		Log.Warn("error reconfiguring logging",
			zap.Error(err),
		)
		return
	}
}