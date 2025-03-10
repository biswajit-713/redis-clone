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

func getEvictionStrategy() EvictionStrategy {
	switch config.EvictionStrategy {
	case "LRU":
		return &EvictFirst{}
	case "LFU":
		return &EvictFirst{}
	default:
		return &EvictFirst{}
	}
}

func evict() {
	strategy := getEvictionStrategy()
	strategy.evict(store)
}
