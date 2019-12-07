package main

import (
	"net/http"
	"os"

	//"github.com/brunovlucena/apps/app-example/cmd/myapp/data"
	"github.com/go-chi/chi"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func main() {
	r := chi.NewRouter()

	// prometheus
	r.Handle("/metrics", promhttp.Handler())

	// add routes
	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("welcome!"))
	})
	listAddr := ":" + os.Getenv("API_CONTAINER_PORT")
	http.ListenAndServe(listAddr, r)
}
