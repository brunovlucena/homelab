package main

import (
	"os"

	"github.com/brunovlucena/mobimeo/apps/app-example/cmd/myapp/router"
	"github.com/sirupsen/logrus"
)

var (
	appName = "myapp"
	r       = router.NewRouter()
)

func main() {
	// Log Config
	logrus.SetFormatter(&logrus.JSONFormatter{})
	// Default is stderr
	logrus.SetOutput(os.Stdout)
	// Will log anything that is info or above (warn, error, fatal, panic).
	logrus.SetLevel(logrus.InfoLevel)
	// App config
	serverAddr := ":" + os.Getenv("API_CONTAINER_PORT")
	// Start App
	r.StartWebServerHTTP(appName, serverAddr)
}
