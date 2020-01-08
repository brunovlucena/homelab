package utils

import (
	"os"

	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

func LogrusSetup() {
	// Setup formater
	logrus.SetFormatter(&logrus.JSONFormatter{})

	// Default is stderr
	logrus.SetOutput(os.Stdout)

	// Will log anything that is info or above (warn, error, fatal, panic).
	logrus.SetLevel(logrus.InfoLevel)
}

func ViperSetup() {
	// Setup path
	viper.SetConfigName("config.yaml")
	viper.AddConfigPath("/app")
	viper.AddConfigPath(".")
	viper.SetConfigType("yaml")

	// Read config
	HandleError(nil, viper.ReadInConfig())
}

func HandleError(result interface{}, err error) (r interface{}) {
	if err != nil {
		logrus.Errorf("Fatal error config file: %s \n", err)
		panic(err)
	}
	return result
}
