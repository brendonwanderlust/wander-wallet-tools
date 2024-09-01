package logger

import (
	"os"
	"time"
	"wander-wallet-tools/utils"

	logger "github.com/sirupsen/logrus"
)

func Init() {
	logger.SetLevel(logger.TraceLevel)
	logger.SetOutput(os.Stdout)
	logger.SetFormatter(&logger.JSONFormatter{
		PrettyPrint:     true,
		TimestampFormat: time.DateTime,
	})
}
func LogInfoLn(message string) {
	logger.Infoln(message)
}

func LogErrorLn(message string, err error) {
	logger.WithField("Error", err.Error()).Errorln(message)
}

func LogFatalLn(message string, err error) {
	logger.WithField("Fatal error", message).Fatalln(utils.IfElse(err == nil, "", err.Error()))
}

func LogInfoWithFields(message string, fields logger.Fields) {
	logger.WithFields(fields).Info(message)
}

func LogErrorWithFields(message string, fields logger.Fields) {
	logger.WithFields(fields).Error(message)
}
