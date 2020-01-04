package utils

import (
	"github.com/sirupsen/logrus"
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
	viper.AddConfigPath("$HOME/.guestbook")
	viper.AddConfigPath(".")
	viper.AddConfigPath("../")
	viper.SetConfigType("yaml")

	// Read config
	HandleError(nil, viper.ReadInConfig())
}

func HandleError(result interface{}, err error) (r interface{}) {
	if err != nil {
		panic(err)
		panic(logrus.Errorf("Fatal error config file: %s \n", err))
	}
	return result
}
