// Package communication handels the translation from http and json-text to
// type-safe golang code. Thus, the actual api-handler doesn't have to bother
// (un-)marshalling json and type-safety.
// A handler-function can be registered for a specific http-method using the
// provided functions with the given names (Get, Post, ...). These functions
// support different types of handler-functions. The raw type of those
// handler-functions is interface{}, since the required specifications can't be
// handled at compile-time in golang. The actual type-pattern is described in
// the type's documentation. When such a handler-function is registered, its
// type is asserted via reflection, so that type-safety is re-established right
// after startup. If the handler-function doesn't match the specified
// type-pattern the registering-function (Get, Post, ...) panics with
// ErrIllegalHandleFunc
package communication

import (
	"encoding/json"
	"errors"
	"net/http"
	"reflect"
	"strconv"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/theMomax/notypo-backend/config"
)

// ErrIllegalHandleFunc is an error thrown, when a function is registered, that
// doesn't meet the requirements issued in the docs above the according
// HandleFunc-type
var ErrIllegalHandleFunc = errors.New("the given handler does not meet the requirements of HandleFunc")

// HandleGetFunc represents a handler-function for a GET request. It has the
// following signature:
//  func(parameters ParameterMap) (status int, response A)
// Where A may be anything json.Marshal can handle
type HandleGetFunc interface{}

// HandlePostFunc represents a handler-function for a POST request. It has the
// following signature:
//  func(request A, parameters ParameterMap) (status int, response B)
// Where A/B may be anything json.Unmarshal/json.Marshal can handle
type HandlePostFunc interface{}

// HandlePutFunc represents a handler-function for a PUT request. It has the
// following signature:
//  func(request A, parameters ParameterMap) (status int)
// Where A may be anything json.Unmarshal can handle
type HandlePutFunc interface{}

// HandleDeleteFunc represents a handler-function for a DELETE request. It has
// the following signature:
//  func(request A, parameters ParameterMap) (status int, response B)
// Where A/B may be anything json.Unmarshal/json.Marshal can handle
type HandleDeleteFunc interface{}

// HandleOptionsFunc represents a handler-function for a OPTIONS request. It has
// the following signature:
//  func(parameters ParameterMap) (status int, response A)
// Where A may be anything json.Marshal can handle
type HandleOptionsFunc interface{}

// ParameterMap contains the parameters of a http-request. The key is the
// parameter's name and the value its value
type ParameterMap map[string]string

var router = mux.NewRouter()

// Serve starts the REST-api and websocket server
func Serve() {
	http.ListenAndServe(config.Server.IP+":"+strconv.Itoa(config.Server.Port), handlers.CORS(
		handlers.AllowedMethods([]string{"GET", "POST", "PUT", "DELETE", "OPTIONS"}),
		handlers.AllowedOrigins(config.Server.AllowedRequestOrigins),
		handlers.AllowedHeaders([]string{"Content-Type"}),
	)(router))
}

// Router returns the router used in this package. This function should only be
// used by tests
func Router() *mux.Router {
	return router
}

// Get registers a handler for the http GET method
func Get(path string, handler HandleGetFunc) {
	// assert, that the given handler meets the requirements of a HandleGetFunc
	hT := reflect.TypeOf(handler)
	hV := reflect.ValueOf(handler)
	if hT.Kind() != reflect.Func {
		panic(ErrIllegalHandleFunc)
	}
	if hT.NumIn() != 1 || hT.NumOut() != 2 {
		panic(ErrIllegalHandleFunc)
	}

	paramsT := hT.In(0)
	if paramsT.Kind() != reflect.Map {
		panic(ErrIllegalHandleFunc)
	}
	if paramsT.Key().Kind() != reflect.String || paramsT.Elem().Kind() != reflect.String {
		panic(ErrIllegalHandleFunc)
	}
	statusT := hT.Out(0)
	if statusT.Kind() != reflect.Int {
		panic(ErrIllegalHandleFunc)
	}
	// register type-safe handler
	router.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		out := hV.Call([]reflect.Value{reflect.ValueOf(mux.Vars(r))})
		response := out[1].Interface()
		bytes, err := json.Marshal(response)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		status := out[0].Interface().(int)
		w.WriteHeader(status)
		w.Write(bytes)
	}).Methods("GET")
}

// Post registers a handler for the http POST method
func Post(path string, handler HandlePostFunc) {
	// assert, that the given handler meets the requirements of a HandlePostFunc
	hT := reflect.TypeOf(handler)
	hV := reflect.ValueOf(handler)
	if hT.Kind() != reflect.Func {
		panic(ErrIllegalHandleFunc)
	}
	if hT.NumIn() != 2 || hT.NumOut() != 2 {
		panic(ErrIllegalHandleFunc)
	}
	requestT := hT.In(0)
	parseRequest := true
	if requestT.String() == "interface {}" || requestT.String() == "*interface {}" {
		parseRequest = false
	}

	paramsT := hT.In(1)
	if paramsT.Kind() != reflect.Map {
		panic(ErrIllegalHandleFunc)
	}
	if paramsT.Key().Kind() != reflect.String || paramsT.Elem().Kind() != reflect.String {
		panic(ErrIllegalHandleFunc)
	}
	statusT := hT.Out(0)
	if statusT.Kind() != reflect.Int {
		panic(ErrIllegalHandleFunc)
	}
	// register type-safe handler
	router.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		var err error
		request := reflect.New(requestT)
		if parseRequest {
			err = json.NewDecoder(r.Body).Decode(request.Interface())
			if err != nil {
				w.WriteHeader(http.StatusBadRequest)
				return
			}
		}
		out := hV.Call([]reflect.Value{request.Elem(), reflect.ValueOf(mux.Vars(r))})
		response := out[1].Interface()
		bytes, err := json.Marshal(response)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		status := out[0].Interface().(int)
		w.WriteHeader(status)
		w.Write(bytes)
	}).Methods("POST")
}

// Put registers a handler for the http PUT method
func Put(path string, handler HandlePutFunc) {
	// assert, that the given handler meets the requirements of a HandlePutFunc
	hT := reflect.TypeOf(handler)
	hV := reflect.ValueOf(handler)
	if hT.Kind() != reflect.Func {
		panic(ErrIllegalHandleFunc)
	}
	if hT.NumIn() != 2 || hT.NumOut() != 1 {
		panic(ErrIllegalHandleFunc)
	}
	requestT := hT.In(0)
	parseRequest := true
	if requestT.String() == "interface {}" || requestT.String() == "*interface {}" {
		parseRequest = false
	}

	paramsT := hT.In(1)
	if paramsT.Kind() != reflect.Map {
		panic(ErrIllegalHandleFunc)
	}
	if paramsT.Key().Kind() != reflect.String || paramsT.Elem().Kind() != reflect.String {
		panic(ErrIllegalHandleFunc)
	}
	statusT := hT.Out(0)
	if statusT.Kind() != reflect.Int {
		panic(ErrIllegalHandleFunc)
	}
	// register type-safe handler
	router.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		var err error
		request := reflect.New(requestT)
		if parseRequest {
			err = json.NewDecoder(r.Body).Decode(request.Interface())
			if err != nil {
				w.WriteHeader(http.StatusBadRequest)
				return
			}
		}
		out := hV.Call([]reflect.Value{request.Elem(), reflect.ValueOf(mux.Vars(r))})
		status := out[0].Interface().(int)
		w.WriteHeader(status)
	}).Methods("PUT")
}

// Delete registers a handler for the http DELETE method
func Delete(path string, handler HandleDeleteFunc) {
	// assert, that the given handler meets the requirements of a HandleDeleteFunc
	hT := reflect.TypeOf(handler)
	hV := reflect.ValueOf(handler)
	if hT.Kind() != reflect.Func {
		panic(ErrIllegalHandleFunc)
	}
	if hT.NumIn() != 2 || hT.NumOut() != 2 {
		panic(ErrIllegalHandleFunc)
	}
	requestT := hT.In(0)
	parseRequest := true
	if requestT.String() == "interface {}" || requestT.String() == "*interface {}" {
		parseRequest = false
	}

	paramsT := hT.In(1)
	if paramsT.Kind() != reflect.Map {
		panic(ErrIllegalHandleFunc)
	}
	if paramsT.Key().Kind() != reflect.String || paramsT.Elem().Kind() != reflect.String {
		panic(ErrIllegalHandleFunc)
	}
	statusT := hT.Out(0)
	if statusT.Kind() != reflect.Int {
		panic(ErrIllegalHandleFunc)
	}
	// register type-safe handler
	router.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		var err error
		request := reflect.New(requestT)
		if parseRequest {
			err = json.NewDecoder(r.Body).Decode(request.Interface())
			if err != nil {
				w.WriteHeader(http.StatusBadRequest)
				return
			}
		}
		out := hV.Call([]reflect.Value{request.Elem(), reflect.ValueOf(mux.Vars(r))})
		response := out[1].Interface()
		bytes, err := json.Marshal(response)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		status := out[0].Interface().(int)
		w.WriteHeader(status)
		w.Write(bytes)
	}).Methods("DELETE")
}

// Options registers a handler for the http OPTIONS method
func Options(path string, handler HandleOptionsFunc) {
	// assert, that the given handler meets the requirements of a HandleOptionsFunc
	hT := reflect.TypeOf(handler)
	hV := reflect.ValueOf(handler)
	if hT.Kind() != reflect.Func {
		panic(ErrIllegalHandleFunc)
	}
	if hT.NumIn() != 1 || hT.NumOut() != 2 {
		panic(ErrIllegalHandleFunc)
	}

	paramsT := hT.In(0)
	if paramsT.Kind() != reflect.Map {
		panic(ErrIllegalHandleFunc)
	}
	if paramsT.Key().Kind() != reflect.String || paramsT.Elem().Kind() != reflect.String {
		panic(ErrIllegalHandleFunc)
	}
	statusT := hT.Out(0)
	if statusT.Kind() != reflect.Int {
		panic(ErrIllegalHandleFunc)
	}
	// register type-safe handler
	router.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		out := hV.Call([]reflect.Value{reflect.ValueOf(mux.Vars(r))})
		response := out[1].Interface()
		bytes, err := json.Marshal(response)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		status := out[0].Interface().(int)
		w.WriteHeader(status)
		w.Write(bytes)
	}).Methods("OPTIONS")
}
