package communication

import (
	"net/http"
	"time"

	"github.com/theMomax/notypo-backend/config"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"github.com/theMomax/notypo-backend/streams"
)

// HandleStreamFunc represents a handler-function for websocket connection. It
// is based on a http GET request. If status is not successful (starting with 2)
// requests are rejected
type HandleStreamFunc func(params map[string]string) (status int, stream streams.Stream)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

// Stream registers a websocket Stream-handler. When a client requests such a
// Stream, a websocket-connection is established. It can be closed by ether
// client or server. The latter closes the connection automatically, when the
// underlying channel (provided by the handler via the streams.Stream interface)
// is closed. The server only sends the Stream's values, when requested. I.e.
// the client must send a JSON-encoded uint value, which represents the number
// of requested streams.Characters.
// The streams.Characters are sent in JSON format. The actual representation
// depends on the underlying streams.StreamSource and how it was initialized
func Stream(path string, handler HandleStreamFunc) {
	router.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		status, stream := handler(mux.Vars(r))
		if (status / 100) != 2 {
			w.WriteHeader(status)
			return
		}
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		requests := make(chan uint, 5)
		closed := make(chan bool, 1)
		go func() {
			for {
				i := uint(0)
				err := conn.ReadJSON(&i)
				if err != nil {
					closed <- true
					close(closed)
					close(requests)
					return
				}
				requests <- i
			}
		}()
	outer:
		for {
			select {
			case <-time.After(config.StreamBase.StreamTimeout):
				conn.Close()
				break outer
			case <-closed:
				conn.Close()
				break outer
			case n := <-requests:
				for i := 0; i < int(n); i++ {
					c, ok := <-stream.Channel()
					if !ok {
						conn.Close()
						break outer
					}
					err := conn.WriteJSON(c)
					if err != nil {
						conn.Close()
						break outer
					}
				}
			}
		}
	})
}
