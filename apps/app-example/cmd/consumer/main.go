package main

import (
	"os"

	"github.com/brunovlucena/mobimeo/apps/app-example/cmd/consumer/config"
	"github.com/brunovlucena/mobimeo/apps/app-example/cmd/consumer/messaging"
	"github.com/sirupsen/logrus"
)

func main() {
	// Vars
	appName := "consumer"

	// Log Config
	logrus.SetFormatter(&logrus.JSONFormatter{})

	// Default is stderr
	logrus.SetOutput(os.Stdout)

	// Will log anything that is info or above (warn, error, fatal, panic).
	logrus.SetLevel(logrus.InfoLevel)

	// App config

	// Start
}

func initializeMessaging() {

	serviceURL := os.Getenv("AMQP_SERVER_URL")
	serviceConfig := os.Getenv("AMQP_CONFIG")

	client := &messaging.MessagingClient{}

	// connect
	client.ConnectToBroker(serviceURL)

	// subscribe
	client.Subscribe(serviceConfig, "topic", appName, config.HandleRefreshEvent)
}
