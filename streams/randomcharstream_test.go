package streams

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

type char rune

func TestCloseFunction(t *testing.T) {
	src := NewRandomCharStreamSource(charslice('a', 'b', 'c', 'd', 'e'))
	s := src.Instance()
	c := s.Channel()
	<-c
	<-c
	<-c
	<-c
	s.Close()
	_, ok := <-c
	assert.False(t, ok)
}

func TestEqualityOfInstances(t *testing.T) {
	src := NewRandomCharStreamSource(charslice('a', 'b', 'c', 'd', 'e'))
	s0 := src.Instance()
	s1 := src.Instance()
	c0 := s0.Channel()
	c1 := s1.Channel()
	for i := 0; i < 100; i++ {
		assert.Equal(t, <-c0, <-c1)
	}
	s0.Close()
	s1.Close()
}

// consume channels the given channel into a slice and returns it
func consume(c chan Character) (s []Character) {
	for {
		v, ok := <-c
		if !ok {
			return
		}
		s = append(s, v)
	}
}

func charslice(elements ...rune) (s []Character) {
	s = make([]Character, len(elements))
	for i, e := range elements {
		s[i] = char(e)
	}
	return
}

func (c char) Rune() rune {
	return rune(c)
}
