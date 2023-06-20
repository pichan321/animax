package animax

import (
	"os"

	"github.com/sirupsen/logrus"
	prefixed "github.com/x-cray/logrus-prefixed-formatter"
)

func GetLogger() (*logrus.Logger){
	logger := logrus.New()
	logger.SetFormatter(&prefixed.TextFormatter{})
	logger.SetOutput(os.Stdout)
	logger.SetFormatter(&logrus.TextFormatter{
		FullTimestamp: true,
		TimestampFormat: "2006-01-02 15:04:05",
		ForceColors: true,
	})
	return logger
}