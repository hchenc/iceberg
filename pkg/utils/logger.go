package utils

import (
	"github.com/sirupsen/logrus"
	"os"
	"time"
)

var (
	logger *logrus.Logger
)

func init() {
	logger = logrus.New()

	logger.Out = os.Stdout

	//logger.Level = logrus.ErrorLevel

	logger.Formatter = &logrus.JSONFormatter{}
}

func GetLogger(fields logrus.Fields) *logrus.Logger {
	logger.WithFields(logrus.Fields{"created": time.Now().UnixNano() / 1e6}).WithFields(fields)
	return logger
}

func GetLoggerEntry() *logrus.Entry {
	return logger.WithFields(logrus.Fields{"created": time.Now().UnixNano() / 1e6})
}
