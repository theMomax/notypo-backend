// Package api contains the implementation of notypo's streaming-api.
// This file is structured differently to golang's standard order. I.e. the file
// is structued by operations. On top, there is the Serve function, that
// registers all operations. The operations are listed below in the same order
// as they are registered in the Serve function. The public constants and types
// relevant to the operation are declared right above the actual function. The
// private ones and further dependencies right below. The actual api is
// documented externally using swagger. The comments in this document provide a
// developers-view
package api

import (
	"net/http"
	"strconv"

	com "github.com/theMomax/notypo-backend/communication"
	"github.com/theMomax/notypo-backend/config"
	"github.com/theMomax/notypo-backend/streams"
)

// Character is the interface, which defines a character for this api
type Character streams.Character

// paths
const (
	PathVersion                    = "/version"
	PathStreamOptions              = "/stream"
	PathCreateStream               = "/stream"
	PathOpenStreamConnection       = "/stream/{id}"
	PathCloseStreamConnection      = "/stream/{id}"
	PathEstablishWebsocketToStream = "/stream/websocket/{id}"
)

// Serve starts the webserver which implements the api specified in this file
func Serve() {
	com.Serve()
}

// Register registers the api-functions specified in this file at the
// http/websocket communication-unit
func Register() {
	com.Get(PathVersion, version)
	com.Get(PathStreamOptions, streamOptions)
	com.Post(PathCreateStream, createStream)
	com.Get(PathOpenStreamConnection, openStream)
	com.Delete(PathCloseStreamConnection, closeStream)
	com.Stream(PathEstablishWebsocketToStream, getStream)
}

// -----------------------------------------------------------------------------
// GET PathVersion
// -----------------------------------------------------------------------------

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
// GET PathStreamOptions
// -----------------------------------------------------------------------------

const (
	// Random StreamSources provide an endless Stream of streams.Characters. The
	// Characters are randomly picked from a given charset
	Random StreamSourceType = "Random"
	// Dictionary StreamSources provide an endless Stream of streams.Characters.
	// The Characters form words, wich are randomly picked from a dictionary's
	// subset, wich only consists of a given charset
	Dictionary StreamSourceType = "Dictionary"
)

// StreamSourceType is a code for a specific type of streams.StreamSource
type StreamSourceType string

// StreamOptionsResponse is a list of all StreamSourceTypes implemented by this
// api-version
type StreamOptionsResponse []StreamSourceType

func streamOptions(params map[string]string) (status int, res StreamOptionsResponse) {
	return http.StatusOK, StreamOptionsResponse{
		Random,
	}
}

// -----------------------------------------------------------------------------
// POST PathCreateStream
// -----------------------------------------------------------------------------

// BasicCharacter is the most basic implementation of streams.Character
type BasicCharacter rune

// StreamSupplierDescription (request) specifies the type and properties of a
// StreamSupplier
type StreamSupplierDescription struct {
	Type    StreamSourceType `json:"type"`
	Charset []BasicCharacter `json:"charset"`
}

// StreamSupplierID (response)
type StreamSupplierID *int64

func createStream(req *StreamSupplierDescription, params map[string]string) (status int, res StreamSupplierID) {
	var source streams.StreamSource
	switch req.Type {
	case Random:
		charset := make([]streams.Character, len(req.Charset))
		for i, c := range req.Charset {
			charset[i] = c
		}
		source = streams.NewRandomCharStreamSource(charset)
	default:
		return http.StatusNotImplemented, nil
	}
	if source == nil {
		return http.StatusBadRequest, nil
	}
	id := streams.Register(source)
	res = &id
	return http.StatusOK, res
}

// Rune returns the character as a rune
func (c BasicCharacter) Rune() rune {
	return rune(c)
}

// -----------------------------------------------------------------------------
// GET PathOpenStreamConnection
// -----------------------------------------------------------------------------

// StreamID (response)
type StreamID *int64

func openStream(params map[string]string) (status int, res StreamID) {
	id, err := strconv.ParseInt(params["id"], 10, 64)
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
// DELETE PathCloseStreamConnection
// -----------------------------------------------------------------------------

func closeStream(req interface{}, params map[string]string) (status int, res interface{}) {
	id, err := strconv.ParseInt(params["id"], 10, 64)
	if err != nil {
		return http.StatusBadRequest, nil
	}
	streams.Close(id)
	return http.StatusOK, nil
}

// -----------------------------------------------------------------------------
// GET/WEBSOCKET PathEstablishWebsocketToStream
// -----------------------------------------------------------------------------

func getStream(params map[string]string) (status int, stream streams.Stream) {
	id, err := strconv.ParseInt(params["id"], 10, 64)
	if err != nil {
		return http.StatusBadRequest, nil
	}
	var ok bool
	stream, ok = streams.Get(id)
	if !ok || stream == nil {
		return http.StatusNotFound, nil
	}
	return http.StatusOK, stream
}
