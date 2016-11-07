package ufwb

import (
	"fmt"
	"github.com/Sirupsen/logrus"
)

func LogAtLevel(level logrus.Level, args ...interface{}) {
	switch level {
	case logrus.DebugLevel:
		logrus.Debug(args)
	case logrus.InfoLevel:
		logrus.Info(args)
	case logrus.WarnLevel:
		logrus.Warn(args)
	case logrus.ErrorLevel:
		logrus.Error(args)
	case logrus.FatalLevel:
		logrus.Fatal(args)
	case logrus.PanicLevel:
		logrus.Panic(args)
	default:
		panic(fmt.Sprintf("Unknown log level: %s", level))
	}
}
