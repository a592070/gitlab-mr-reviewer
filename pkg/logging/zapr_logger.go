package logging

import (
	"github.com/go-logr/logr"
	"github.com/go-logr/zapr"
	"github.com/pkg/errors"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"log"
)

type ZaprLogger struct {
	logr.Logger
	core *zap.Logger
}

func NewZaprLogger(isReleaseMode bool, logLevel string) (*ZaprLogger, error) {
	var c zap.Config
	level, err := zapcore.ParseLevel(logLevel)
	if err != nil {
		log.Printf("[NewZaprLogger]Invalid log level, using default: info. err: %v\n", err.Error())
		level = zapcore.InfoLevel
	}

	if isReleaseMode {
		c = zap.NewProductionConfig()
	} else {
		c = zap.NewDevelopmentConfig()
		c.EncoderConfig.FunctionKey = "F"
	}
	c.Level = zap.NewAtomicLevelAt(level)
	logger, err := c.Build(
		zap.AddCallerSkip(1), // traverse call depth for more useful log lines
		zap.AddCaller())
	if err != nil {
		return nil, errors.Wrap(err, "[NewZaprLogger]failed to build logger")
	}

	zaprLogger := zapr.NewLogger(logger)
	return &ZaprLogger{
		zaprLogger,
		logger,
	}, nil
}

func (l ZaprLogger) WithError(err error) ZaprLogger {
	l.Logger = l.Logger.WithValues("error", err.Error())
	return l
}

func (l ZaprLogger) Info(msg string) {
	l.Logger.Info(msg)
}

func (l ZaprLogger) Debug(msg string) {
	l.Logger.V(1).Info(msg)
}

func (l ZaprLogger) Warn(msg string) {
	l.Logger.V(-1).Info(msg)
}

func (l ZaprLogger) GetCore() *zap.Logger {
	return l.core
}
