package logger

import (
	"os"
	"strings"

	"github.com/sirupsen/logrus"
)

var log *logrus.Logger

func InitLogger(environment string) {
	log = logrus.New()

	log.SetOutput(os.Stdout)

	if strings.ToLower(environment) == "development" {
		log.SetLevel(logrus.DebugLevel)
	} else {
		log.SetLevel(logrus.InfoLevel)
	}

	if strings.ToLower(environment) == "development" {
		log.SetFormatter(&logrus.TextFormatter{
			FullTimestamp:   true,
			TimestampFormat: "2006-01-02 15:04:05",
			ForceColors:     true,
			PadLevelText:    true,
		})
	} else {
		log.SetFormatter(&logrus.JSONFormatter{
			TimestampFormat: "2006-01-02T15:04:05.000Z",
			PrettyPrint:     false,
		})
	}

	log.Info("Logger initialized successfully")
}
func GetLogger() *logrus.Logger {
	if log == nil {
		InitLogger("development")
	}
	return log
}
func WithField(key string, value interface{}) *logrus.Entry {
	return GetLogger().WithField(key, value)
}
func WithFields(fields logrus.Fields) *logrus.Entry {
	return GetLogger().WithFields(fields)
}
func WithError(err error) *logrus.Entry {
	return GetLogger().WithError(err)
}
func Debug(args ...interface{}) {
	GetLogger().Debug(args...)
}
func Info(args ...interface{}) {
	GetLogger().Info(args...)
}
func Warn(args ...interface{}) {
	GetLogger().Warn(args...)
}
func Error(args ...interface{}) {
	GetLogger().Error(args...)
}
func Fatal(args ...interface{}) {
	GetLogger().Fatal(args...)
}
func Debugf(format string, args ...interface{}) {
	GetLogger().Debugf(format, args...)
}
func Infof(format string, args ...interface{}) {
	GetLogger().Infof(format, args...)
}
func Warnf(format string, args ...interface{}) {
	GetLogger().Warnf(format, args...)
}
func Errorf(format string, args ...interface{}) {
	GetLogger().Errorf(format, args...)
}
func Fatalf(format string, args ...interface{}) {
	GetLogger().Fatalf(format, args...)
}
