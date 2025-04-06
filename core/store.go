package core

import (
	"strings"
	"time"

	"github.com/diceclone/config"
)

var store map[string]*Obj

// TODO - make the attributes private and provide public wrapper method to access the attributes
// TODO - Obj has TypeEncoding field, as of now it will support only integer, raw string and embedded string

func init() {
	store = make(map[string]*Obj)
}

func Put(key string, value *Obj) {
	if len(store) >= config.KeysLimit {
		evict()
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
		return true
	}

	return false
}

func (o Obj) HasExpired() bool {
	return o.ValidTill != -1 && o.ValidTill < int(time.Now().Unix())
}

func (o Obj) TtlSet() bool {
	return o.ValidTill != -1
}

func IterateStore() <-chan *KeyValuePair {
	ch := make(chan *KeyValuePair)

	go func() {
		defer close(ch)
		for key, obj := range store {
			ch <- &KeyValuePair{Key: key, Value: obj}
		}
	}()

	return ch
}

type KeyValuePair struct {
	Key   string
	Value *Obj
}
