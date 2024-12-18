package core

var store map[string]*Obj

type Obj struct {
	Value interface{}
}

func init() {
	store = make(map[string]*Obj)
}

func Put(key string, value *Obj) {
	store[key] = value
}

func NewObj(value interface{}) *Obj {
	return &Obj{Value: value}
}
