package core

import (
	"github.com/diceclone/config"
)

type EvictionStrategy interface {
	evict(store map[string]*Obj)
}

type EvictFirst struct{}

func (e *EvictFirst) evict(store map[string]*Obj) {
	for k := range store {
		delete(store, k)
		break
	}
}

type EvictRandom struct{}

func (e *EvictRandom) evict(store map[string]*Obj) {
	evictionSize := EvictionSize()
	for k := range store {
		delete(store, k)
		evictionSize--
		if evictionSize == 0 {
			break
		}
	}
}

func getEvictionStrategy() EvictionStrategy {
	switch config.EvictionStrategy {
	case "LRU":
		return &EvictFirst{}
	case "LFU":
		return &EvictFirst{}
	case "EVICT_RANDOM":
		return &EvictRandom{}
	default:
		return &EvictFirst{}
	}
}

func Evict() {
	if len(store) >= config.KeysLimit {
		strategy := getEvictionStrategy()
		strategy.evict(store)
	}

	// when the key size reaches KeysLimit, evict 40% of the keys
	// it is inefficient to calculate the store everytime Evict() is called
	// hence the store size must be pre-computed
	var evictionSize = EvictionSize()
	if evictionSize != 0 {
		strategy := getEvictionStrategy()
		strategy.evict(store)
	}
}
