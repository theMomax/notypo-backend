package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"runtime"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"github.com/stretchr/testify/assert"
	com "github.com/theMomax/notypo-backend/communication"
	"github.com/theMomax/notypo-backend/config"
)

var r *mux.Router

func init() {
	config.IsTest = true
	config.ConfigPath = "config.ini"
	config.Load(nil)
	config.Version = "TESTVERSION"
	config.GitCommit = "TESTGITCOMMIT"
	config.BuildTime = "TESTBUILDTIME"
	config.IsTest = false

	Register()
	r = com.Router()
}

func TestVersion(t *testing.T) {
	ngr := runtime.NumGoroutine()

	body := bytes.NewBuffer(make([]byte, 0))
	req, _ := http.NewRequest("GET", "/version", body)
	resp := httptest.NewRecorder()
	r.ServeHTTP(resp, req)
	assert.Equal(t, 200, resp.Code)
	assert.Equal(t, jsons(VersionResponse{
		Version:   config.Version,
		GitCommit: config.GitCommit,
		BuildTime: config.BuildTime,
	}), resp.Body.String())

	config.IsTest = true
	resp = httptest.NewRecorder()
	r.ServeHTTP(resp, req)
	assert.Equal(t, 503, resp.Code)
	config.IsTest = false

	assert.Equal(t, ngr, runtime.NumGoroutine())
}

func TestCreateStream501(t *testing.T) {
	ngr := runtime.NumGoroutine()

	body := bytes.NewBuffer(make([]byte, 0))
	json.NewEncoder(body).Encode(StreamSupplierDescription{
		Type: "other",
		Charset: []rune{
			'a', 'b', 'c',
		},
	})
	req, _ := http.NewRequest("POST", "/stream", body)
	resp := httptest.NewRecorder()
	r.ServeHTTP(resp, req)
	assert.Equal(t, 501, resp.Code)

	assert.Equal(t, ngr, runtime.NumGoroutine())
}

func TestOpenStream404(t *testing.T) {
	ngr := runtime.NumGoroutine()

	body := bytes.NewBuffer(make([]byte, 0))
	req, _ := http.NewRequest("GET", "/stream/1", body)
	resp := httptest.NewRecorder()
	r.ServeHTTP(resp, req)
	assert.Equal(t, 404, resp.Code)

	assert.Equal(t, ngr, runtime.NumGoroutine())
}

func TestStreamProcedureWebsocket404(t *testing.T) {
	config.StreamBase.StreamTimeout = time.Hour
	config.StreamBase.SupplierTimeout = time.Hour
	s := httptest.NewServer(r)
	defer s.Close()
	ngr := runtime.NumGoroutine()

	body := bytes.NewBuffer(make([]byte, 0))
	json.NewEncoder(body).Encode(StreamSupplierDescription{
		Type: Random,
		Charset: []rune{
			'a', 'b', 'c',
		},
	})

	req, _ := http.NewRequest("POST", "/stream", body)
	resp := httptest.NewRecorder()
	r.ServeHTTP(resp, req)
	assert.Equal(t, 200, resp.Code)
	var streamID int64
	assert.NoError(t, json.Unmarshal(resp.Body.Bytes(), &streamID))

	body = bytes.NewBuffer(make([]byte, 0))
	req, _ = http.NewRequest("GET", "/stream/"+strconv.FormatInt(streamID, 10), body)
	resp = httptest.NewRecorder()
	r.ServeHTTP(resp, req)
	assert.Equal(t, 200, resp.Code)
	var connectionID int64
	assert.NoError(t, json.Unmarshal(resp.Body.Bytes(), &connectionID))

	_, _, err := websocket.DefaultDialer.Dial("ws"+strings.TrimPrefix(s.URL, "http")+"/stream/websocket/1", nil)
	assert.Error(t, err)

	body = bytes.NewBuffer(make([]byte, 0))
	req, _ = http.NewRequest("DELETE", "/stream/"+strconv.FormatInt(connectionID, 10), body)
	resp = httptest.NewRecorder()
	r.ServeHTTP(resp, req)
	assert.Equal(t, 200, resp.Code)

	<-time.After(20 * time.Millisecond)
	assert.Equal(t, ngr, runtime.NumGoroutine())
}

func TestCloseStreamNoEffect(t *testing.T) {
	ngr := runtime.NumGoroutine()

	body := bytes.NewBuffer(make([]byte, 0))
	req, _ := http.NewRequest("DELETE", "/stream/1", body)
	resp := httptest.NewRecorder()
	r.ServeHTTP(resp, req)
	assert.Equal(t, 200, resp.Code)

	assert.Equal(t, ngr, runtime.NumGoroutine())
}

func TestStreamProcedureWithoutConnection(t *testing.T) {
	config.StreamBase.StreamTimeout = time.Hour
	config.StreamBase.SupplierTimeout = 0
	ngr := runtime.NumGoroutine()

	body := bytes.NewBuffer(make([]byte, 0))
	json.NewEncoder(body).Encode(StreamSupplierDescription{
		Type: Random,
		Charset: []rune{
			'a', 'b', 'c',
		},
	})

	req, _ := http.NewRequest("POST", "/stream", body)
	resp := httptest.NewRecorder()
	r.ServeHTTP(resp, req)
	assert.Equal(t, 200, resp.Code)
	var streamID int64
	assert.NoError(t, json.Unmarshal(resp.Body.Bytes(), &streamID))

	<-time.After(20 * time.Millisecond)
	assert.Equal(t, ngr, runtime.NumGoroutine())
}

func TestStreamProcedureWithoutConnectionClosing(t *testing.T) {
	config.StreamBase.StreamTimeout = 0
	config.StreamBase.SupplierTimeout = time.Hour
	for i := 0; i < 100; i++ {
		ngr := runtime.NumGoroutine()

		body := bytes.NewBuffer(make([]byte, 0))
		json.NewEncoder(body).Encode(StreamSupplierDescription{
			Type: Random,
			Charset: []rune{
				'a', 'b', 'c',
			},
		})

		req, _ := http.NewRequest("POST", "/stream", body)
		resp := httptest.NewRecorder()
		r.ServeHTTP(resp, req)
		assert.Equal(t, 200, resp.Code)
		var streamID int64
		assert.NoError(t, json.Unmarshal(resp.Body.Bytes(), &streamID))

		body = bytes.NewBuffer(make([]byte, 0))
		req, _ = http.NewRequest("GET", "/stream/"+strconv.FormatInt(streamID, 10), body)
		resp = httptest.NewRecorder()
		r.ServeHTTP(resp, req)
		assert.Equal(t, 200, resp.Code)
		var connectionID int64
		assert.NoError(t, json.Unmarshal(resp.Body.Bytes(), &connectionID))

		<-time.After(20 * time.Millisecond)
		assert.Equal(t, ngr, runtime.NumGoroutine())
	}
}

func TestStreamProcedureWithoutWebsocketClosing(t *testing.T) {
	config.StreamBase.StreamTimeout = 10 * time.Millisecond
	config.StreamBase.SupplierTimeout = time.Hour
	s := httptest.NewServer(r)
	defer s.Close()
	for i := 0; i < 550; i++ {
		ngr := runtime.NumGoroutine()

		body := bytes.NewBuffer(make([]byte, 0))
		json.NewEncoder(body).Encode(StreamSupplierDescription{
			Type: Random,
			Charset: []rune{
				'a', 'b', 'c',
			},
		})

		req, _ := http.NewRequest("POST", "/stream", body)
		resp := httptest.NewRecorder()
		r.ServeHTTP(resp, req)
		assert.Equal(t, 200, resp.Code)
		var streamID int64
		assert.NoError(t, json.Unmarshal(resp.Body.Bytes(), &streamID))

		body = bytes.NewBuffer(make([]byte, 0))
		req, _ = http.NewRequest("GET", "/stream/"+strconv.FormatInt(streamID, 10), body)
		resp = httptest.NewRecorder()
		r.ServeHTTP(resp, req)
		assert.Equal(t, 200, resp.Code)
		var connectionID int64
		assert.NoError(t, json.Unmarshal(resp.Body.Bytes(), &connectionID))

		ws, _, err := websocket.DefaultDialer.Dial("ws"+strings.TrimPrefix(s.URL, "http")+"/stream/websocket/"+strconv.FormatInt(connectionID, 10), nil)
		assert.NoError(t, err)

		assert.NoError(t, ws.WriteJSON(uint(3)))
		for j := 0; j < 3; j++ {
			var r rune
			assert.NoError(t, ws.ReadJSON(&r))
		}

		body = bytes.NewBuffer(make([]byte, 0))
		req, _ = http.NewRequest("DELETE", "/stream/"+strconv.FormatInt(connectionID, 10), body)
		resp = httptest.NewRecorder()
		r.ServeHTTP(resp, req)
		assert.Equal(t, 200, resp.Code)

		<-time.After(30 * time.Millisecond)
		assert.Equal(t, ngr, runtime.NumGoroutine())
	}
}

func TestStreamProcedureWithoutWebsocketClosingButConnectionClosing(t *testing.T) {
	config.StreamBase.StreamTimeout = time.Hour
	config.StreamBase.SupplierTimeout = time.Hour
	s := httptest.NewServer(r)
	defer s.Close()
	for i := 0; i < 100; i++ {
		ngr := runtime.NumGoroutine()

		body := bytes.NewBuffer(make([]byte, 0))
		json.NewEncoder(body).Encode(StreamSupplierDescription{
			Type: Random,
			Charset: []rune{
				'a', 'b', 'c',
			},
		})

		req, _ := http.NewRequest("POST", "/stream", body)
		resp := httptest.NewRecorder()
		r.ServeHTTP(resp, req)
		assert.Equal(t, 200, resp.Code)
		var streamID int64
		assert.NoError(t, json.Unmarshal(resp.Body.Bytes(), &streamID))

		body = bytes.NewBuffer(make([]byte, 0))
		req, _ = http.NewRequest("GET", "/stream/"+strconv.FormatInt(streamID, 10), body)
		resp = httptest.NewRecorder()
		r.ServeHTTP(resp, req)
		assert.Equal(t, 200, resp.Code)
		var connectionID int64
		assert.NoError(t, json.Unmarshal(resp.Body.Bytes(), &connectionID))

		ws, _, err := websocket.DefaultDialer.Dial("ws"+strings.TrimPrefix(s.URL, "http")+"/stream/websocket/"+strconv.FormatInt(connectionID, 10), nil)
		assert.NoError(t, err)

		assert.NoError(t, ws.WriteJSON(uint(3)))
		for j := 0; j < 3; j++ {
			var r rune
			assert.NoError(t, ws.ReadJSON(&r))
		}

		body = bytes.NewBuffer(make([]byte, 0))
		req, _ = http.NewRequest("DELETE", "/stream/"+strconv.FormatInt(connectionID, 10), body)
		resp = httptest.NewRecorder()
		r.ServeHTTP(resp, req)
		assert.Equal(t, 200, resp.Code)

		assert.NoError(t, ws.WriteJSON(uint(3)))

		<-time.After(20 * time.Millisecond)
		assert.Equal(t, ngr, runtime.NumGoroutine())
	}
}

func TestStreamProcedure(t *testing.T) {
	config.StreamBase.StreamTimeout = time.Hour
	config.StreamBase.SupplierTimeout = time.Hour
	s := httptest.NewServer(r)
	defer s.Close()
	for i := 0; i < 100; i++ {
		ngr := runtime.NumGoroutine()

		body := bytes.NewBuffer(make([]byte, 0))
		json.NewEncoder(body).Encode(StreamSupplierDescription{
			Type: Random,
			Charset: []rune{
				'a', 'b', 'c',
			},
		})

		req, _ := http.NewRequest("POST", "/stream", body)
		resp := httptest.NewRecorder()
		r.ServeHTTP(resp, req)
		assert.Equal(t, 200, resp.Code)
		var streamID int64
		assert.NoError(t, json.Unmarshal(resp.Body.Bytes(), &streamID))

		body = bytes.NewBuffer(make([]byte, 0))
		req, _ = http.NewRequest("GET", "/stream/"+strconv.FormatInt(streamID, 10), body)
		resp = httptest.NewRecorder()
		r.ServeHTTP(resp, req)
		assert.Equal(t, 200, resp.Code)
		var connectionID int64
		assert.NoError(t, json.Unmarshal(resp.Body.Bytes(), &connectionID))

		ws, _, err := websocket.DefaultDialer.Dial("ws"+strings.TrimPrefix(s.URL, "http")+"/stream/websocket/"+strconv.FormatInt(connectionID, 10), nil)
		assert.NoError(t, err)

		assert.NoError(t, ws.WriteJSON(uint(3)))
		for j := 0; j < 3; j++ {
			var r rune
			assert.NoError(t, ws.ReadJSON(&r))
		}
		ws.Close()

		body = bytes.NewBuffer(make([]byte, 0))
		req, _ = http.NewRequest("DELETE", "/stream/"+strconv.FormatInt(connectionID, 10), body)
		resp = httptest.NewRecorder()
		r.ServeHTTP(resp, req)
		assert.Equal(t, 200, resp.Code)

		<-time.After(20 * time.Millisecond)
		assert.Equal(t, ngr, runtime.NumGoroutine())
	}
}

func jsons(i interface{}) string {
	b, _ := json.Marshal(i)
	return string(b)
}
