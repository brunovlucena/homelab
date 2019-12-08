package main

import (
	"os"

	"github.com/brunovlucena/mobimeo/apps/app-example/cmd/myapp/router"
)

var (
	appName = "myapp"
	r       = router.NewRouter()
)

func main() {
	serverAddr := ":" + os.Getenv("API_CONTAINER_PORT")
	// Start
	r.StartWebServerHTTP(appName, serverAddr)
}
