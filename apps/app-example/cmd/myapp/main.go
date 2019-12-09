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
	// Config
	// Will log anything that is info or above (warn, error, fatal, panic).
	logrus.SetLevel(logrus.InfoLevel)
	serverAddr := ":" + os.Getenv("API_CONTAINER_PORT")
	// End
	// Start App
	r.StartWebServerHTTP(appName, serverAddr)
}
