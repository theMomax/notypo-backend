// Package streams contains the logic for managing and generating Streams. The
// streambase.go file contains the rather generic management, while the other
// files in this package contain implementations of StreamSources. A Stream is
// a endless source of Characters, that automatically pipes those Characters
// into a read-only channel
package streams

import (
	"errors"
	"log"
	"math/rand"
	"sync"
	"time"

	"github.com/theMomax/notypo-backend/config"
)

// errors
var (
	ErrNoSuchSupplier = errors.New("there is no supplier registered under the given id")
)

// StreamSupplier represents a registered StreamSource
type StreamSupplier interface {
	StreamSource
	// ID returns a unique identifier
	ID() int64
}

// StreamSource represents a source of Streams
type StreamSource interface {
	// Instance returns a Stream. Each Stream must generate the same output in
	// the same order
	Instance() UnregisteredStream
}

// Stream is a wrapper for a registered channel of Characters. The
// Stream automalltically channels Characters into the Channel until
// it is closed
type Stream interface {
	UnregisteredStream
	// ID returns a unique identifier
	ID() int64
}

// UnregisteredStream is a wrapper for a channel of Characters. The
// UnregisteredStream automalltically channels Characters into the Channel until
// it is closed
type UnregisteredStream interface {
	// Channel returns the actual channel of Characters
	Channel() <-chan Character
	// Close closes Channel(). It may not panic, if called multiple times
	Close()
}

// Character is the required type for the input-streams
type Character interface {
	// Rune returns the character's utf-8 representation
	Rune() rune
}

type streamSupplier struct {
	StreamSource
	id      int64
	connect chan bool
}

type streamWrapper struct {
	UnregisteredStream
	id         int64
	supplierID int64
}

var suppliers map[int64]*streamSupplier
var suplm sync.RWMutex

var streams map[int64]*streamWrapper
var strm sync.RWMutex

func init() {
	suppliers = make(map[int64]*streamSupplier)
	suplm = sync.RWMutex{}
	streams = make(map[int64]*streamWrapper)
	strm = sync.RWMutex{}
	rand.Seed(time.Now().UTC().UnixNano())
}

// Register registrates a StreamSource, so that it is available via the returned
// id as a StreamSupplier afterwards. The StreamSupplier is unregistered, when
// ether no Instance is requested from this source within
// config.StreamBase.SupplierTimeout, or at least one Instance has been opened
// and all Instances were closed again since then
func Register(source StreamSource) (id int64) {
	id = rand.Int63()
	s := &streamSupplier{
		StreamSource: source,
		id:           id,
		// controls deletion: true means and additional connection was openend,
		// false means a connection was closed
		connect: make(chan bool),
	}
	writeSupplier(id, s)
	go manageUnregistration(s)
	return
}

// Open returns the id of a new Instance of the StreamSupplier with the given id
// or returns an ErrNoSuchSupplier, if the id is invalid. The Stream is closed
// at latest config.StreamBase.StreamTimeout after it was opened
func Open(supplierID int64) (streamID int64, err error) {
	defer func() {
		err := recover()
		if err != nil {
			log.Println(err)
		}
	}()
	supl := readSupplier(supplierID)
	if supl == nil {
		return 0, ErrNoSuchSupplier
	}
	streamID = rand.Int63()
	s := &streamWrapper{
		UnregisteredStream: supl.Instance(),
		id:                 streamID,
		supplierID:         supplierID,
	}
	writeStream(streamID, s)
	// recover from send-to-closed-supl.connect-channel panic, in case the
	// supplier was unregistered since the initial check
	defer func() {
		e := recover()
		if e != nil {
			err = ErrNoSuchSupplier
		}
	}()
	supl.connect <- true
	// close on timeout
	time.AfterFunc(config.StreamBase.StreamTimeout, func() {
		Close(s.ID())
	})
	return
}

// Get returns the Stream with the given id. Get returns !ok if there is
// no such stream
func Get(streamID int64) (stream Stream, ok bool) {
	defer func() {
		err := recover()
		if err != nil {
			log.Println(err)
		}
	}()
	return readStream(streamID)
}

// Close closes and deletes the Stream with the given id
func Close(streamID int64) {
	defer func() {
		err := recover()
		if err != nil {
			log.Println(err)
		}
	}()
	s, ok := readStream(streamID)
	if ok && s != nil {
		su := readSupplier(s.supplierID)
		if su != nil {
			su.connect <- false
		}
		s.Close()
		deleteStream(streamID)
	}
}

func manageUnregistration(supplier *streamSupplier) {
	defer func() {
		err := recover()
		if err != nil {
			log.Println(err)
		}
	}()
	connections := 0
	timeout := time.After(config.StreamBase.SupplierTimeout)
	for {
		select {
		case <-timeout:
			close(supplier.connect)
			deleteSupplier(supplier.ID())
			return
		case c := <-supplier.connect:
			if c {
				connections++
				// reset timeout
				timeout = time.After(config.StreamBase.SupplierTimeout)
			} else {
				connections--
				if connections == 0 {
					close(supplier.connect)
					deleteSupplier(supplier.ID())
					return
				}
			}
		}
	}
}

func readSupplier(id int64) (supplier *streamSupplier) {
	suplm.RLock()
	supplier = suppliers[id]
	suplm.RUnlock()
	return
}

func writeSupplier(id int64, supplier *streamSupplier) {
	suplm.Lock()
	suppliers[id] = supplier
	suplm.Unlock()
	return
}

func deleteSupplier(id int64) {
	suplm.Lock()
	delete(suppliers, id)
	suplm.Unlock()
}

func readStream(id int64) (stream *streamWrapper, ok bool) {
	strm.RLock()
	stream, ok = streams[id]
	strm.RUnlock()
	return
}

func writeStream(id int64, stream *streamWrapper) {
	strm.Lock()
	streams[id] = stream
	strm.Unlock()
	return
}

func deleteStream(id int64) {
	strm.Lock()
	delete(streams, id)
	strm.Unlock()
}

func (s *streamSupplier) ID() int64 {
	return s.id
}

func (s *streamWrapper) ID() int64 {
	return s.id
}
