package main

import "go.uber.org/zap"

var (
	logger       *zap.SugaredLogger
	loggerConfig = zap.Config{
		Level:            zap.NewAtomicLevelAt(zap.ErrorLevel),
		Development:      false,
		DisableCaller:    true,
		Encoding:         "console",
		EncoderConfig:    zap.NewDevelopmentEncoderConfig(),
		OutputPaths:      []string{"stdout"},
		ErrorOutputPaths: []string{"stderr"},
	}
)

func init() {
	basicLogger, err := loggerConfig.Build()
	if err != nil {
		panic("cannot open logger")
	}
	logger = basicLogger.Sugar()
	defer logger.Sync()
}

func setLogLevel(level string) {
	switch level {
	case "error":
		loggerConfig.Level.SetLevel(zap.ErrorLevel)
	case "warn":
		loggerConfig.Level.SetLevel(zap.WarnLevel)
	case "info":
		loggerConfig.Level.SetLevel(zap.InfoLevel)
	case "verbose":
		loggerConfig.Level.SetLevel(zap.DebugLevel)
	default:
		panic("wrong log level, only error, warn, info and verbose are support")
	}
}
