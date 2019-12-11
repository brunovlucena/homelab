package router

import (
	"context"
	"fmt"
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
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/sirupsen/logrus"
	"golang.org/x/net/http2"
)

var (
	repo      repository.Repository
	histogram = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Name: "create_duration_seconds",
		Help: "Time taken to create a config",
	}, []string{"code"})
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
	// log error
	utils.LogErr(err)
	// prometheus
	prometheus.Register(histogram)
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
	// log error
	utils.LogErr(err)
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
		// render errors
		if err != nil {
			render.Render(w, r, ErrRender(err))
			return
		}
		// save config
		ctx := context.WithValue(r.Context(), "config", config)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// FindsAll returns all configs.
func FindAll(w http.ResponseWriter, r *http.Request) {
	// start of create
	start := time.Now()
	// get configs
	configs, err := repo.FindAll()
	// render errors
	if err != nil {
		code := http.StatusUnprocessableEntity // 422
		// prometheus: observe error
		duration := time.Since(start)
		observe(duration, code)
		// render
		render.Render(w, r, ErrRender(err))
		return
	}
	// Status Created
	code := http.StatusCreated
	// prometheus: observe created
	duration := time.Since(start)
	logrus.WithFields(logrus.Fields{
		"cmd":      "FindAll",
		"duration": duration,
	}).Info("Record created!")
	observe(duration, code)
	// render
	render.Status(r, code)
	render.RenderList(w, r, NewConfigListResponse(configs))
}

// Create creates a new config.
func Create(w http.ResponseWriter, r *http.Request) {
	// start of create
	start := time.Now()
	// convert post data into json format
	cr := ConfigRequest{}
	err := cr.Bind(r)
	// check error
	logrus.WithFields(logrus.Fields{
		"cmd": "Bind",
	}).Error(err.Error())
	// persist data in the database
	c, err := repo.Create(&cr.Config)
	// render errors
	if err != nil {
		// Status Error
		code := http.StatusUnprocessableEntity // 422
		// prometheus: observe error
		duration := time.Since(start)
		observe(duration, code)
		observe(duration, code)
		// render
		render.Render(w, r, ErrRender(err))
		return
	}
	// Status Created
	code := http.StatusCreated
	// prometheus: observe created
	duration := time.Since(start)
	logrus.WithFields(logrus.Fields{
		"cmd":      "create",
		"duration": duration,
	}).Info("Record created!")
	observe(duration, code)
	// render
	render.Status(r, code)
	render.Render(w, r, NewConfigResponse(c))
}

// Find returns the specified config.
func Find(w http.ResponseWriter, r *http.Request) {
	// get from context
	config := r.Context().Value("config").(*data.Config)
	render.Render(w, r, NewConfigResponse(config))
	render.Status(r, http.StatusFound)
}

// Update updates the specified Config.
func Update(w http.ResponseWriter, r *http.Request) {
	// start of update
	start := time.Now()
	// convert post data into json format
	cr := ConfigRequest{}
	err := cr.Bind(r)
	// check error
	logrus.WithFields(logrus.Fields{
		"cmd": "Bind",
	}).Error(err.Error())
	// update record
	c, err := repo.Update(&cr.Config)
	// render errors
	if err != nil {
		// Status Error
		code := http.StatusUnprocessableEntity // 422
		// prometheus: observe error
		duration := time.Since(start)
		observe(duration, code)
		// render
		render.Render(w, r, ErrRender(err))
		return
	}
	render.Status(r, http.StatusOK)
	render.Render(w, r, NewConfigResponse(c))
}

// Delete removes the specified Config.
func Delete(w http.ResponseWriter, r *http.Request) {
	// get data from context
	config := r.Context().Value("config").(*data.Config)
	// removes from database
	_, err := repo.Remove(config.Data["name"].(string))
	// check errors
	if err != nil {
		logrus.Error(err)
		render.Render(w, r, ErrRender(err))
		return
	}
	render.Status(r, http.StatusFound)
	render.Render(w, r, NewConfigResponse(config))
}

// Search returns the Configs data for a matching config.
// Query:   /metadata.{key}={value}
func Search(w http.ResponseWriter, r *http.Request) {

}

func observe(duration time.Duration, code int) {
	histogram.WithLabelValues(fmt.Sprintf("%d", code)).Observe(duration.Seconds())
}
