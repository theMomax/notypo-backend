package streams

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/theMomax/notypo-backend/config"
)

func init() {
	config.IsTest = true
	config.ConfigPath = "config.ini"
	config.Load(nil)
}

func TestUnregisterOnStreamSupplierTimeout(t *testing.T) {
	config.StreamBase.SupplierTimeout = 50 * time.Millisecond
	config.StreamBase.StreamTimeout = 50 * time.Second
	src := NewRandomCharStreamSource(charslice('a', 'b', 'c', 'd', 'e'))
	id := Register(src)
	time.Sleep(60 * time.Millisecond)
	_, err := Open(id)
	assert.Equal(t, ErrNoSuchSupplier, err)
}

func TestUnregisterOnAllStreamsClosed(t *testing.T) {
	config.StreamBase.SupplierTimeout = 50 * time.Second
	config.StreamBase.StreamTimeout = 50 * time.Second
	src := NewRandomCharStreamSource(charslice('a', 'b', 'c', 'd', 'e'))
	id := Register(src)
	sid, _ := Open(id)
	Close(sid)
	_, err := Open(id)
	assert.Equal(t, ErrNoSuchSupplier, err)
}

func TestStreamTimeout(t *testing.T) {
	config.StreamBase.SupplierTimeout = 50 * time.Second
	config.StreamBase.StreamTimeout = 50 * time.Millisecond
	src := NewRandomCharStreamSource(charslice('a', 'b', 'c', 'd', 'e'))
	id := Register(src)
	Open(id)
	time.Sleep(60 * time.Millisecond)
	_, err := Open(id)
	assert.Equal(t, ErrNoSuchSupplier, err)
}

func TestStreamClosing(t *testing.T) {
	config.StreamBase.SupplierTimeout = 50 * time.Second
	config.StreamBase.StreamTimeout = 50 * time.Second
	src := NewRandomCharStreamSource(charslice('a', 'b', 'c', 'd', 'e'))
	id := Register(src)
	sid, _ := Open(id)
	Close(sid)
	assert.Nil(t, Get(sid))
}
