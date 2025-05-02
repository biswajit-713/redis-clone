package core

import (
	"sort"
	"time"

	"github.com/diceclone/config"
)

var evictionPool []*Obj

func InitializePool() {
	evictionPool = make([]*Obj, 0)
}

func canBeEvicted(obj *Obj) bool {

	if len(evictionPool) < config.EVICTION_POOL_SIZE {
		evictionPool = append(evictionPool, obj)
		arrangeEvictionPool()
		return false
	}

	logger.Printf("Eviction pool size: %d\n", len(evictionPool))
	if isIdleForLonger(obj) {
		evictionPool = evictionPool[1:]
		evictionPool = append(evictionPool, obj)
		arrangeEvictionPool()
		logger.Printf("%s is older, candidate for eviction.\n", obj.Value)
		return true
	}

	return false
}

func arrangeEvictionPool() {
	sort.Slice(evictionPool, func(i, j int) bool {
		return evictionPool[j].LastAccessedAt < evictionPool[i].LastAccessedAt
	})
}

func isIdleForLonger(obj *Obj) bool {
	latOfWorstCandidate := evictionPool[len(evictionPool)-1].LastAccessedAt
	latOfCurrentCandidate := obj.LastAccessedAt

	logger.Printf("lat of obj: %d, lat of last element: %d, lat of first element: %d\n", latOfCurrentCandidate, latOfWorstCandidate, evictionPool[0].LastAccessedAt)

	return idleTimeOf(latOfCurrentCandidate) > idleTimeOf(latOfWorstCandidate)
}

func idleTimeOf(lat uint32) uint32 {
	currentClock := time.Now().Unix() & 0x00FFFFFF

	if int64(lat) < currentClock {
		return lat
	} else {
		return uint32(currentClock + int64(0x00FFFFFF-lat))
	}
}
