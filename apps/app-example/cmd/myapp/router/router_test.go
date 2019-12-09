package router

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/brunovlucena/mobimeo/apps/app-example/cmd/myapp/utils"
	. "github.com/smartystreets/goconvey/convey"
)

var (
	configs []map[string]interface{}
	r       *MyRouter
)

func init() {
	// initialise router
	r = NewRouter()
	// server in background
	go r.StartWebServerHTTP("test", "0.0.0.0:8000")
	// load json
	utils.LoadJson("router_test.json", &configs)
}

func TestCreate(t *testing.T) {

	// create pod-1z
	Convey("Given a HTTP request for /configs to create pod-1z", t, func() {
		//jsonStr, _ := json.Marshal(configs[0])
		// name is missing

		jsonStr := `{"name": "pod-1z","metadata": {"monitoring": {"enabled": "true"},"limits": {"cpu": {"enabled": "false","value": "300m"}}}}`

		res := bytes.NewBuffer([]byte(jsonStr))
		req := httptest.NewRequest("POST", "/configs", res)
		resp := httptest.NewRecorder()

		Convey("When the request is handled by the Router", func() {
			r.Mux.ServeHTTP(resp, req)
			Convey("Then the response should be a 201", func() {
				So(resp.Code, ShouldEqual, http.StatusCreated)
			})
		})
	})
}

func TestDelete(t *testing.T) {

	// delete pod-1z
	Convey("Given a HTTP request for /configs/pod-1z", t, func() {
		req := httptest.NewRequest("DELETE", "/configs/pod-1z", nil)
		resp := httptest.NewRecorder()

		Convey("When the request is handled by the Router", func() {
			r.Mux.ServeHTTP(resp, req)
			Convey("Then the response should be a 302", func() {
				So(resp.Code, ShouldEqual, http.StatusFound)
			})
		})
	})
}

// TODO: Test fails only because render.Status is not working
func TestFindAll(t *testing.T) {
	// find
	Convey("Given a HTTP request for /configs to find configs", t, func() {
		req := httptest.NewRequest("GET", "/configs", nil)
		resp := httptest.NewRecorder()

		Convey("When the request is handled by the Router", func() {
			r.Mux.ServeHTTP(resp, req)
			Convey("Then the response should be a 302", func() {
				So(resp.Code, ShouldEqual, http.StatusFound)
			})
		})
	})
}

// TODO: Test fails only because render.Status is not working
func TestFind(t *testing.T) {

	// find
	Convey("Given a HTTP request for /configs/pod-2", t, func() {
		req := httptest.NewRequest("GET", "/configs/pod-2", nil)
		resp := httptest.NewRecorder()

		Convey("When the request is handled by the Router", func() {
			r.Mux.ServeHTTP(resp, req)
			Convey("Then the response should be a 302", func() {
				So(resp.Code, ShouldEqual, http.StatusFound)
			})
		})
	})
}

func TestUpdate(t *testing.T) {

	jsonStr := `{"name": "pod-11","metadata": {"monitoring": {"enabled": "true"},"limits": {"cpu": {"enabled": "false","value": "300m"}}}}`

	Convey("Given a HTTP request for /configs/pod-11", t, func() {

		res := bytes.NewBuffer([]byte(jsonStr))
		req := httptest.NewRequest("PUT", "/configs/pod-11", res)
		resp := httptest.NewRecorder()

		Convey("When the request is handled by the Router", func() {
			r.Mux.ServeHTTP(resp, req)

			Convey("Then the response should be a 200", func() {
				So(resp.Code, ShouldEqual, http.StatusOK)
			})
		})
	})
}

func TestSearch(t *testing.T) {

}
