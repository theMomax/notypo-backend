// Package api contains the implementation of notypo's streaming-api.
// This file is structured differently to golang's standard order. I.e. the file
// is structued by operations. On top, there is the Serve function, that
// registers all operations. The operations are listed below in the same order
// as they are registered in the Serve function. The public constants and types
// relevant to the operation are declared right above the actual function. The
// private ones right below. The actual api is documented externally using
// swagger. The comments in this document provide a developers-view
package api

import (
	"net/http"
	"strconv"

	com "github.com/theMomax/notypo-backend/communication"
	"github.com/theMomax/notypo-backend/config"
	"github.com/theMomax/notypo-backend/streams"
)

// Serve starts the webserver which implements the api specified in this file
func Serve() {
	com.Get("/version", version)
	com.Post("/stream", createStream)
	com.Get("/stream/{id}", openStream)
	com.Delete("/stream/{id}", closeStream)
	com.Stream("(/stream/websocket/{id}", getStream)

	com.Serve()
}

// VersionResponse holds general information about this program
type VersionResponse struct {
	Version   string `json:"version"`
	GitCommit string `json:"git_commit"`
	BuildTime string `json:"build_time"`
}

func version(params map[string]string) (status int, res *VersionResponse) {
	if config.IsTest {
		return http.StatusServiceUnavailable, nil
	}
	return http.StatusOK, &VersionResponse{
		Version:   config.Version,
		GitCommit: config.GitCommit,
		BuildTime: config.BuildTime,
	}
}

// -----------------------------------------------------------------------------
// POST /stream
// -----------------------------------------------------------------------------

const (
	// Random StreamSources provide an endless Stream of streams.Characters. The
	// Characters are randomly picked from a given charset
	Random StreamSourceType = iota
	// Dictionary StreamSources provide an endless Stream of streams.Characters.
	// The Characters form words, wich are randomly picked from a dictionary's
	// subset, wich only consists of a given charset
	Dictionary
)

// StreamSourceType is a code for a specific type of streams.StreamSource
type StreamSourceType int

// StreamSupplierDescription (request) specifies the type and properties of a
// StreamSupplier
type StreamSupplierDescription struct {
	// Example: 1
	Type    StreamSourceType `json:"type"`
	Charset []rune           `json:"charset"`
}

// StreamSupplierID (response)
type StreamSupplierID *uint64

func createStream(req *StreamSupplierDescription, params map[string]string) (status int, res StreamSupplierID) {
	var source streams.StreamSource
	switch req.Type {
	case Random:
		charset := make([]streams.Character, len(req.Charset))
		for i, c := range req.Charset {
			charset[i] = character(c)
		}
		source = streams.NewRandomCharStreamSource(charset)
	default:
		return http.StatusNotImplemented, nil
	}
	id := streams.Register(source)
	res = &id
	return http.StatusOK, res
}

type character rune

func (c character) Rune() rune {
	return rune(c)
}

// -----------------------------------------------------------------------------
// GET /stream/{id}
// -----------------------------------------------------------------------------

// StreamID (response)
type StreamID *uint64

func openStream(params map[string]string) (status int, res StreamID) {
	id, err := strconv.ParseUint(params["id"], 10, 64)
	if err != nil {
		return http.StatusBadRequest, nil
	}
	streamID, err := streams.Open(id)
	if err != nil {
		return http.StatusNotFound, nil
	}
	return http.StatusOK, &streamID
}

// -----------------------------------------------------------------------------
// DELETE /stream/{id}
// -----------------------------------------------------------------------------

func closeStream(req interface{}, params map[string]string) (status int, res interface{}) {
	id, err := strconv.ParseUint(params["id"], 10, 64)
	if err != nil {
		return http.StatusBadRequest, nil
	}
	streams.Close(id)
	return http.StatusOK, nil
}

// -----------------------------------------------------------------------------
// GET /stream/websocket/{id}
// -----------------------------------------------------------------------------

func getStream(params map[string]string) (status int, stream streams.Stream) {
	id, err := strconv.ParseUint(params["id"], 10, 64)
	if err != nil {
		return http.StatusBadRequest, nil
	}
	stream = streams.Get(id)
	if stream == nil {
		return http.StatusNotFound, nil
	}
	return http.StatusOK, stream
}
