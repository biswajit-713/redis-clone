package core

import (
	"strings"
	"time"

	"github.com/diceclone/config"
)

var store map[string]*Obj
var keysCount int = 0

// TODO - make the attributes private and provide public wrapper method to access the attributes
// TODO - Obj has TypeEncoding field, as of now it will support only integer, raw string and embedded string

func init() {
	store = make(map[string]*Obj)
}

func Put(key string, value *Obj) {
	// takes care of evicting policy
	Evict()

	if !exists(key) {
		keysCount++
	}
	store[strings.ToUpper(key)] = value

}

func Get(k string) *Obj {
	if v, ok := store[strings.ToUpper(k)]; ok {
		return v
	}
	return nil
}

func Delete(k string) bool {
	if _, ok := store[strings.ToUpper(k)]; ok {
		delete(store, strings.ToUpper(k))
		keysCount--
		return true
	}

	return false
}

func ClearDB() {
	store = make(map[string]*Obj)
	keysCount = 0
}

func exists(k string) bool {
	_, ok := store[strings.ToUpper(k)]
	return ok
}

func (o Obj) HasExpired() bool {
	return o.ValidTill != -1 && o.ValidTill < int(time.Now().Unix())
}

func (o Obj) TtlSet() bool {
	return o.ValidTill != -1
}

type KeyValuePair struct {
	Key   string
	Value *Obj
}

func IterateStore() <-chan *KeyValuePair {
	ch := make(chan *KeyValuePair)
	keysCount = 0
	go func() {
		defer close(ch)
		for key, obj := range store {
			keysCount++
			ch <- &KeyValuePair{Key: key, Value: obj}
		}
	}()

	return ch
}

func EvictionSize() int {
	if keysCount < config.KeysLimit {
		return 0
	}
	return int(float64(config.KeysLimit) * config.EvictionRatio)
}

func KeyspaceSize() int {
	return keysCount
}

func computeKeyspaceSize() {
	keysCount = len(store)
}
