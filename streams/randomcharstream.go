package streams

import (
	"math/rand"
	"time"

	"github.com/tevino/abool"
)

type randomCharStreamSource struct {
	seed    int64
	charset []Character
}

type basicUnregisteredCharStream struct {
	channel chan Character
	rand    *rand.Rand
	closed  *abool.AtomicBool
}

// NewRandomCharStreamSource creates a StreamSource, which pipes the same
// random sequence of Characters into each of its Instances. The Characters are
// taken from the given charset. If the charset is nil or empty, nil is returned
func NewRandomCharStreamSource(charset []Character) StreamSource {
	if charset == nil || len(charset) < 1 {
		return nil
	}
	return &randomCharStreamSource{
		seed:    time.Now().UTC().UnixNano(),
		charset: charset,
	}
}

func (r *randomCharStreamSource) Instance() UnregisteredStream {
	b := &basicUnregisteredCharStream{
		channel: make(chan Character),
		rand:    rand.New(rand.NewSource(r.seed)),
		closed:  abool.New(),
	}
	// pipe random Characters into the channel, until Close() is called
	go func() {
		for !b.closed.IsSet() {
			c := r.charset[b.rand.Intn(len(r.charset))]
			b.channel <- c
		}
		close(b.channel)
	}()
	return b
}

func (b *basicUnregisteredCharStream) Channel() <-chan Character {
	return b.channel
}

func (b *basicUnregisteredCharStream) Close() {
	b.closed.Set()
}
