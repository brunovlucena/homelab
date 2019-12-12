package router

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
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
		Name: "task_duration_seconds",
		Help: "Time taken to performa a task",
	}, []string{"code", "function"})
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
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"cmd": "init",
		}).Error(err.Error())
	}

	// prometheus
	prometheus.Register(histogram)
}

// StartWebServerHTTP starts the App
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
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"cmd": "StartWebServerHTTP",
		}).Error(err.Error())
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
		// start tracking
		start := time.Now()

		sk := strings.Split(r.URL.Path, "/") // E.g route: /configs/foo/bar/
		configName := sk[2]                  // configName: foo

		config, err := repo.Find(configName) // gets from repository
		duration := time.Since(start)
		// end tracking

		// render errors
		if err != nil {
			render.Render(w, r, ErrRender(err))
			return
		} else {
			// prometheus: observe error
			code := http.StatusUnprocessableEntity
			observe(duration, code, "findall")

			// log
			logrus.WithFields(logrus.Fields{
				"cmd":      "Find",
				"duration": duration,
				"code":     code,
			}).Debug("Records found!")
		}

		// prometheus
		code := http.StatusFound
		observe(duration, code, "find")

		// save config
		ctx := context.WithValue(r.Context(), "config", config)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// FindsAll returns all configs.
func FindAll(w http.ResponseWriter, r *http.Request) {
	// start tracking
	start := time.Now()

	// get configs
	configs, err := repo.FindAll()

	duration := time.Since(start)
	// end tracking

	// render errors
	if err != nil {
		// prometheus: observe error
		code := http.StatusUnprocessableEntity
		observe(duration, code, "findall")

		// log
		logrus.WithFields(logrus.Fields{
			"cmd":      "FindAll",
			"duration": duration,
			"code":     code,
		}).Debug("Records not found!")

		// render
		render.Render(w, r, ErrRender(err))
		return
	}

	// prometheus
	code := http.StatusFound
	observe(duration, code, "findall")

	// log
	logrus.WithFields(logrus.Fields{
		"cmd":      "FindAll",
		"duration": duration,
		"code":     code,
	}).Debug("Records found!")

	// render
	render.Status(r, code)
	render.RenderList(w, r, NewConfigListResponse(configs))
}

// Create creates a new config.
func Create(w http.ResponseWriter, r *http.Request) {
	// start tracking
	start := time.Now()

	// convert post data into json format
	cr := ConfigRequest{}
	err := cr.Bind(r)

	// check error
	if err != nil {
		// return error
		render.Render(w, r, ErrRender(err))
		return
	}

	// persist data in the database
	c, err := repo.Create(&cr.Config)
	duration := time.Since(start)

	// render errors
	if err != nil {
		// prometheus: observe error
		code := http.StatusUnprocessableEntity
		observe(duration, code, "create")

		// render
		render.Render(w, r, ErrRender(err))
		return
	}

	// prometheus
	code := http.StatusCreated
	observe(duration, code, "create")

	// log
	logrus.WithFields(logrus.Fields{
		"cmd":      "Create",
		"duration": duration,
		"code":     code,
	}).Info("Record created!")

	// render
	render.Status(r, code)
	render.Render(w, r, NewConfigResponse(c))
}

// Find returns the specified config.
func Find(w http.ResponseWriter, r *http.Request) {
	// get from context
	config := r.Context().Value("config").(*data.Config)

	// render
	render.Render(w, r, NewConfigResponse(config))
	render.Status(r, http.StatusFound)
}

// Update updates the specified Config.
func Update(w http.ResponseWriter, r *http.Request) {
	// start tracking
	start := time.Now()

	// convert post data into json format
	cr := ConfigRequest{}
	err := cr.Bind(r)

	// check error
	if err != nil {
		// return error
		render.Render(w, r, ErrRender(err))
		return
	}

	// update record
	c, err := repo.Update(&cr.Config)
	duration := time.Since(start)
	// end tracking

	// render errors
	if err != nil {
		// prometheus: observe error
		code := http.StatusUnprocessableEntity
		observe(duration, code, "update")

		// render
		render.Render(w, r, ErrRender(err))
		return
	}

	// prometheus
	code := http.StatusFound
	observe(duration, code, "update")

	// log
	logrus.WithFields(logrus.Fields{
		"cmd":      "Update",
		"duration": duration,
		"code":     code,
	}).Info("Record updated!")

	// render
	render.Status(r, code)
	render.Render(w, r, NewConfigResponse(c))
}

// Delete removes the specified Config.
func Delete(w http.ResponseWriter, r *http.Request) {
	// start tracking
	start := time.Now()

	// get config from context
	config := r.Context().Value("config").(*data.Config)

	// removes from database
	_, err := repo.Remove(config.Data["name"].(string))
	duration := time.Since(start)
	// end tracking

	// render errors
	if err != nil {
		// prometheus: observe error
		code := http.StatusUnprocessableEntity
		observe(duration, code, "delete")

		// render
		render.Render(w, r, ErrRender(err))
		return
	}

	// prometheus
	code := http.StatusFound
	observe(duration, code, "delete")

	// log
	logrus.WithFields(logrus.Fields{
		"cmd":      "Delete",
		"duration": duration,
		"code":     code,
	}).Info("Record removed!")

	// render
	render.Status(r, code)
	render.Render(w, r, NewConfigResponse(config))
}

// Search returns the Configs data for a matching config.
// Query:   /metadata.{key}={value}
func Search(w http.ResponseWriter, r *http.Request) {
	// start tracking
	start := time.Now()
	// get params
	params, _ := url.ParseQuery(r.URL.String())

	// get configs
	configs, err := repo.Search(params)
	duration := time.Since(start)
	// end tracking

	// check errors
	if err != nil {
		// prometheus: observe error
		code := http.StatusUnprocessableEntity
		observe(duration, code, "search")

		// render error
		render.Render(w, r, ErrRender(err))
		return
	}

	// prometheus
	code := http.StatusFound
	observe(duration, code, "search")

	// log
	logrus.WithFields(logrus.Fields{
		"cmd":      "Search",
		"duration": duration,
		"code":     code,
	}).Info("Record(s) found!")

	// render result
	render.RenderList(w, r, NewConfigListResponse(configs))
}

// Helper Funcion for Prometheus
func observe(duration time.Duration, code int, task string) {
	histogram.WithLabelValues(fmt.Sprintf("%d", code), task).Observe(duration.Seconds())
}
