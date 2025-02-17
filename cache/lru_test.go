package cache

import (
	"container/list"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

// Test using container/list for tracking recent usage
func TestList(t *testing.T) {
	l := list.New()
	l.PushBack(1)
	l.PushBack(2)
	e3 := l.PushBack(3)

	l.MoveToFront(e3)
	fmt.Println(l.Front().Value)

	l.Remove(e3)
	fmt.Println(l.Front().Value)
}

func TestNoTTL(t *testing.T) {
	c := NewLRUCache(10)
	c.Set("b1", "k1", []byte("v1"), Options{})
	value, err := c.Get("b1", "k1", Options{})
	assert.NoError(t, err)
	assert.Equal(t, []byte("v1"), value)

	_, err = c.Get("b1", "k2", Options{})
	assert.Error(t, err)
}

func TestCapacity(t *testing.T) {
	c := NewLRUCache(3)
	c.Set("b1", "k1", []byte("v1"), Options{})
	c.Set("b1", "k2", []byte("v2"), Options{})
	c.Set("b1", "k3", []byte("v3"), Options{})

	_, err := c.Get("b1", "k1", Options{})
	assert.Error(t, err)

	v, err := c.Get("b1", "k2", Options{})
	assert.NoError(t, err)
	assert.Equal(t, []byte("v2"), v)
}
