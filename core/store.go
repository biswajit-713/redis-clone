package core

import "strings"

var store map[string]*Obj

type Obj struct {
	Value     interface{}
	ValidTill int
}

func init() {
	store = make(map[string]*Obj)
}

func Put(key string, value *Obj) {
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

func NewObj(value interface{}, validTill int) *Obj {
	return &Obj{
		Value:     value,
		ValidTill: validTill,
	}
}
