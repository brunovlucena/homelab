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

	// create
	Convey("Given a HTTP request for /configs to create pod-1r", t, func() {
		//jsonStr, _ := json.Marshal(configs[0])
		// name is missing

		jsonStr := `{"name": "pod-1r","metadata": {"monitoring": {"enabled": "true"},"limits": {"cpu": {"enabled": "false","value": "300m"}}}}`

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

	//// create
	//Convey("Given a HTTP request for /configs to create pod-2r", t, func() {
	//jsonStr, _ := json.Marshal(configs[1])
	//req := httptest.NewRequest("POST", "/configs", bytes.NewBuffer(jsonStr))
	//resp := httptest.NewRecorder()

	//Convey("When the request is handled by the Router", func() {
	//r.Mux.ServeHTTP(resp, req)
	//Convey("Then the response should be a 201", func() {
	//So(resp.Code, ShouldEqual, 201)
	//})
	//})
	//})

	//Convey("Given a HTTP request for /configs to create pod-3r", t, func() {
	//jsonStr, _ := json.Marshal(configs[2])
	//req := httptest.NewRequest("POST", "/configs", bytes.NewBuffer(jsonStr))
	//resp := httptest.NewRecorder()

	//Convey("When the request is handled by the Router", func() {
	//r.Mux.ServeHTTP(resp, req)
	//Convey("Then the response should be a 201", func() {
	//So(resp.Code, ShouldEqual, 201)
	//})
	//})
	//})
}

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

func TestFind(t *testing.T) {

	Convey("Given a HTTP request for /configs/pod-1", t, func() {
		req := httptest.NewRequest("GET", "/configs/pod-1", nil)
		resp := httptest.NewRecorder()

		Convey("When the request is handled by the Router", func() {
			r.Mux.ServeHTTP(resp, req)
			Convey("Then the response should be a 302", func() {
				So(resp.Code, ShouldEqual, http.StatusFound)
			})
		})
	})
}

func TestDelete(t *testing.T) {

	//// delete pod-1b
	//Convey("Given a HTTP request for /configs/pod-1r", t, func() {
	//req := httptest.NewRequest("DELETE", "/configs/pod-1r", nil)
	//resp := httptest.NewRecorder()

	//Convey("When the request is handled by the Router", func() {
	//r.Mux.ServeHTTP(resp, req)
	//Convey("Then the response should be a 302", func() {
	//So(resp.Code, ShouldEqual, 302)
	//})
	//})
	//})

	//// delete pod-2b
	//Convey("Given a HTTP request for /configs/pod-2r", t, func() {
	//req := httptest.NewRequest("DELETE", "/configs/pod-2r", nil)
	//resp := httptest.NewRecorder()

	//Convey("When the request is handled by the Router", func() {
	//r.Mux.ServeHTTP(resp, req)
	//Convey("Then the response should be a 302", func() {
	//So(resp.Code, ShouldEqual, 302)
	//})
	//})
	//})

	//// delete pod-3b
	//Convey("Given a HTTP request for /configs/pod-3r", t, func() {
	//req := httptest.NewRequest("DELETE", "/configs/pod-3r", nil)
	//resp := httptest.NewRecorder()

	//Convey("When the request is handled by the Router", func() {
	//r.Mux.ServeHTTP(resp, req)
	//Convey("Then the response should be a 302", func() {
	//So(resp.Code, ShouldEqual, 302)
	//})
	//})
	//})

	//// Case: delete non-existent config
	//Convey("Given a HTTP request for /configs/idonotexist", t, func() {
	//req := httptest.NewRequest("DELETE", "/configs/idonotexist", nil)
	//resp := httptest.NewRecorder()

	//Convey("When the request is handled by the Router", func() {
	//r.Mux.ServeHTTP(resp, req)

	//Convey("Then the response should be a 422", func() {
	//So(resp.Code, ShouldEqual, 422)
	//})
	//})
	//})
}

func TestUpdate(t *testing.T) {

	//Convey("Given a HTTP request for /configs/pod-3", t, func() {

	//jsonStr, _ := json.Marshal(configs[3])
	//req := httptest.NewRequest("PUT", "/configs/pod-3", bytes.NewBuffer(jsonStr))
	//resp := httptest.NewRecorder()

	//Convey("When the request is handled by the Router", func() {
	//r.Mux.ServeHTTP(resp, req)

	//Convey("Then the response should be a 200", func() {
	//So(resp.Code, ShouldEqual, 200)
	//})
	//})
	//})
}

func TestSearch(t *testing.T) {

}
