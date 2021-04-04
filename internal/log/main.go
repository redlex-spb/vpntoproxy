package log

import (
	"github.com/natefinch/lumberjack"
	"github.com/sirupsen/logrus"
	"io"
	"os"
	"vpntoproxy/internal/config"
)

func SetDefaultSettings() {
	conf := config.Get()

	if conf.Basic.Debug {
		logrus.SetLevel(logrus.DebugLevel)
	} else {
		logrus.SetLevel(logrus.WarnLevel)
	}

	logger := &lumberjack.Logger{
		Filename:   "logs/log.log",
		MaxSize:    conf.Log.MaxSize,
		MaxBackups: conf.Log.MaxBackups,
		MaxAge:     conf.Log.MaxAge,
		Compress:   conf.Log.Compress,
	}

	switch conf.Log.Mode {
	case "file":
		logrus.SetOutput(logger)
	case "multi":
		logrus.SetOutput(io.MultiWriter(logger, os.Stdout))
	default:
		logrus.SetFormatter(&logrus.TextFormatter{
			DisableLevelTruncation: false,
			ForceColors:            true,
			PadLevelText:           true,
		})
		logrus.SetOutput(os.Stdout)
	}
}
