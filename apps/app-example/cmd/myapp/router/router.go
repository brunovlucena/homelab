package router

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/brunovlucena/mobimeo/apps/app-example/cmd/myapp/data"
	"github.com/brunovlucena/mobimeo/apps/app-example/cmd/myapp/repository"
	"github.com/brunovlucena/mobimeo/apps/app-example/cmd/myapp/utils"
	"github.com/go-chi/chi"
	"github.com/go-chi/render"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/sirupsen/logrus"
	"golang.org/x/net/http2"
)

var (
	repo repository.Repository
)

func init() {
	var err error
	dType := os.Getenv("DATABASE_TYPE")
	dHost := os.Getenv("DATABASE_HOST")
	dPort := os.Getenv("DATABASE_PORT")
	dUser := os.Getenv("DATABASE_USER")
	dPass := os.Getenv("DATABASE_PASS")
	dbName := os.Getenv("DATABASE_NAME")
	repo, err = repository.NewRepository(dType, dHost, dPort, dUser, dPass, dbName)
	if err != nil {
		logrus.Infoln(err)
	}
}

func (r *MyRouter) StartWebServerHTTP(appName, serverAddr string) {
	if serverAddr == ":" {
		serverAddr = ":8000" // falls back to default port
	}

	srv := &http.Server{
		Addr:         serverAddr,
		Handler:      r.Mux,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	// Use HTTP2
	err := http2.ConfigureServer(srv, &http2.Server{})
	if err != nil {
		logrus.Infoln("Couldn't upgrade to http2: ", err.Error())
	}
	// setup
	r.setupRoutes()
	// start listening
	logrus.Infof("Starting %v on 0.0.0.0%s", appName, serverAddr)
	logrus.Fatalln(srv.ListenAndServe())
}

func (r *MyRouter) setupRoutes() {
	// add healthcheck
	r.Mux.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		w.WriteHeader(http.StatusOK)
		b, err := ioutil.ReadFile("router/welcome.html")
		utils.LogErr(err)
		welcome := string(b)
		w.Write([]byte(welcome))
	})
	// add healthcheck
	r.Mux.Get("/ping", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("pong"))
	})
	// prometheus
	r.Mux.Handle("/metrics", promhttp.Handler())
	// add routes
	r.Mux.Route("/configs", func(r chi.Router) {
		r.Get("/", FindAll)
		r.Post("/", Create)
		r.Route("/{name}", func(r chi.Router) {
			r.Use(ConfigCtx) // Loads a config in the request's Context
			r.Get("/", Find)
			r.Put("/", Update)
			r.Patch("/", Update)
			r.Delete("/", Delete)
		})
	})
	r.Mux.Get("/search", Search) // GET /search?metadata.{key}={value}
}

// ConfigCtx middleware is used to load an Config object from
// the URL parameters passed through as the request. In case
// the Config could not be found.
func ConfigCtx(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sk := strings.Split(r.URL.Path, "/") // E.g route: /configs/foo/bar/
		configName := sk[2]                  // configName: foo
		config, err := repo.Find(configName) // gets from repository
		if err != nil {
			render.Render(w, r, ErrRender(err))
			return
		}
		ctx := context.WithValue(r.Context(), "config", config)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// FindsAll returns all configs.
func FindAll(w http.ResponseWriter, r *http.Request) {

}

// Create creates a new config.
func Create(w http.ResponseWriter, r *http.Request) {
	// convert post data into json format
	var configJson map[string]interface{}
	err := json.NewDecoder(r.Body).Decode(&configJson)
	utils.LogErr(err)
	// persist data into database
	// create
	config := data.Config{Data: configJson}
	err = repo.Create(&config)
	if err != nil {
		render.Render(w, r, ErrRender(err))
		return
	}
}

// Find returns the specified config.
func Find(w http.ResponseWriter, r *http.Request) {
	// get from context
	config := r.Context().Value("config").(*data.Config)
	render.Status(r, http.StatusOK)
	render.Render(w, r, NewConfigResponse(config))
}

// Update updates the specified Config.
func Update(w http.ResponseWriter, r *http.Request) {
	// convert post data into json format
	var configJson map[string]interface{}
	err := json.NewDecoder(r.Body).Decode(&configJson)
	utils.LogErr(err)
	config := data.Config{Data: configJson}
	// update
	err = repo.Update(&config)
	if err != nil {
		render.Render(w, r, ErrRender(err))
		return
	}
}

// Delete removes the specified Config.
func Delete(w http.ResponseWriter, r *http.Request) {
	// convert post data into json format
	var configJson map[string]interface{}
	err := json.NewDecoder(r.Body).Decode(&configJson)
	utils.LogErr(err)
	config := data.Config{Data: configJson}
	// removes from database
	_, err = repo.Remove(config.Data["name"])
	if err != nil {
		render.Render(w, r, ErrRender(err))
		return
	}
}

// Search returns the Configs data for a matching config.
// Query:   /metadata.{key}={value}
func Search(w http.ResponseWriter, r *http.Request) {

}
