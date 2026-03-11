package main

import (
	"fmt"
	"sync"
)

type opType int

const (
	opSet opType = iota
	opGet
	opDel
	opLen
)

type response[V any] struct {
	value V
	ok    bool
	len   int
}

type command[K comparable, V any] struct {
	op    opType
	key   K
	value V
	resp  chan response[V]
}

type SafeCache[K comparable, V any] struct {
	cmd chan command[K, V]
	wg  sync.WaitGroup
}

func NewSafeCache[K comparable, V any]() *SafeCache[K, V] {
	c := &SafeCache[K, V]{
		cmd: make(chan command[K, V], 128),
	}
	c.wg.Add(1)
	go c.worker()
	return c
}

func (c *SafeCache[K, V]) worker() {
	defer c.wg.Done()
	data := make(map[K]V)

	for cmd := range c.cmd {
		switch cmd.op {
		case opSet:
			data[cmd.key] = cmd.value
			close(cmd.resp)

		case opGet:
			val, ok := data[cmd.key]
			cmd.resp <- response[V]{value: val, ok: ok}
			close(cmd.resp)

		case opDel:
			delete(data, cmd.key)
			close(cmd.resp)

		case opLen:
			cmd.resp <- response[V]{len: len(data)}
			close(cmd.resp)
		}
	}
}

func (c *SafeCache[K, V]) Set(key K, value V) {
	resp := make(chan response[V], 1)
	c.cmd <- command[K, V]{
		op:    opSet,
		key:   key,
		value: value,
		resp:  resp,
	}
	<-resp
}

func (c *SafeCache[K, V]) Get(key K) (V, bool) {
	respCh := make(chan response[V], 1)
	c.cmd <- command[K, V]{
		op:   opGet,
		key:  key,
		resp: respCh,
	}

	r := <-respCh
	return r.value, r.ok
}

func (c *SafeCache[K, V]) Delete(key K) {
	resp := make(chan response[V], 1)
	c.cmd <- command[K, V]{
		op:   opDel,
		key:  key,
		resp: resp,
	}

	<-resp
}

func (c *SafeCache[K, V]) Len() int {
	respCh := make(chan response[V], 1)
	c.cmd <- command[K, V]{
		op:   opLen,
		resp: respCh,
	}

	r := <-respCh
	return r.len
}

func (c *SafeCache[K, V]) Close() {
	close(c.cmd)
	c.wg.Wait()
}

func main() {
	cache := NewSafeCache[string, int]()
	defer cache.Close()

	cache.Set("a", 10)
	cache.Set("b", 20)

	v, ok := cache.Get("a")
	fmt.Println("a--->", v, ok)

	cache.Delete("b")
	fmt.Println("len --->", cache.Len())

}
