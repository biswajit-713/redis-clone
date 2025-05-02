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
	evictionSize := evictionSize()
	for k := range store {
		delete(store, k)
		evictionSize--
		if evictionSize == 0 {
			break
		}
	}
	computeKeyspaceSize()
}

type EvictLru struct{}

func (e *EvictLru) evict(store map[string]*Obj) {

	// TODO - the keys with highest idle time must expire
	// first compute the current clock
	// over a sample size, for each key check if it's idle time is worse than the idle time of the
	// last item in the eviction pool array, if it so, kick it out

	logger.Println("Eviction strategy: LRU")

	for evictionSize() > 0 {
		keysDeleted := 0
		for k := range store {
			if keysDeleted >= config.SAMPLE_SIZE {
				break
			}
			if canBeEvicted(store[k]) {
				Delete(k)
			}
			keysDeleted++
		}
	}
}

func getEvictionStrategy() EvictionStrategy {
	switch config.EVICTION_STRATEGY {
	case "EVICT_LRU":
		return &EvictLru{}
	case "LFU":
		return &EvictFirst{}
	case "EVICT_RANDOM":
		return &EvictRandom{}
	default:
		return &EvictFirst{}
	}
}

func Evict() {
	// when the key size reaches KeysLimit, evict 40% of the keys
	// it is inefficient to calculate the store everytime Evict() is called
	// hence the store size must be pre-computed
	var evictionSize = evictionSize()
	logger.Printf("Eviction triggered: %d keys to be evicted\n", evictionSize)
	if evictionSize != 0 {
		strategy := getEvictionStrategy()
		strategy.evict(store)
	}
}

func evictionSize() int {
	size := KeyspaceSize()
	if size < config.KEYS_LIMIT {
		return 0
	}
	return int(float64(config.KEYS_LIMIT) * config.EVICTION_RATIO)
}
