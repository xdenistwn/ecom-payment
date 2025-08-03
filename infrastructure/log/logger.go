package log

import "github.com/sirupsen/logrus"

var Logger *logrus.Logger

func SetupLogger() {
	log := logrus.New()

	log.SetFormatter(&logrus.TextFormatter{
		ForceColors:   true,
		FullTimestamp: true,
	})

	log.Info("Logged initiated using logrus!")
	Logger = log
}
