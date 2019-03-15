package streams

import (
	"math/rand"
	"time"
)

type randomCharStreamSource struct {
	seed    int64
	charset []Character
}

type basicUnregisteredCharStream struct {
	channel chan Character
	rand    *rand.Rand
	done    chan bool
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
		done:    make(chan bool),
	}
	// pipe random Characters into the channel, until Close() is called
	go func() {
	outer:
		for {
			c := r.charset[b.rand.Intn(len(r.charset))]
			select {
			case b.channel <- c:
			case <-b.done:
				close(b.done)
				break outer
			}
		}
		close(b.channel)
	}()
	return b
}

func (b *basicUnregisteredCharStream) Channel() <-chan Character {
	return b.channel
}

func (b *basicUnregisteredCharStream) Close() {
	b.done <- true
}
