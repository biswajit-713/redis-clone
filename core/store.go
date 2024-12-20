package core

var store map[string]*Obj

type Obj struct {
	Value     interface{}
	ValidTill int
}

func init() {
	store = make(map[string]*Obj)
}

func Put(key string, value *Obj) {
	store[key] = value
}

func Get(k string) *Obj {
	if v, ok := store[k]; ok {
		return v
	}
	return nil
}

func NewObj(value interface{}, validTill int) *Obj {
	return &Obj{
		Value:     value,
		ValidTill: validTill,
	}
}
