package main

import (
	"math"
	"sync"
)

type ChairRideCache struct {
	*ChairLocation
	Distance int64
}

// 新規追加管理
var globalMu = sync.RWMutex{}
// 固有操作管理
var chairMu = make(map[string]*sync.RWMutex)

// chair_idと最新位置情報と走行距離を保持する
var chairRideCacheMap = make(map[string]*ChairRideCache)


func CacheChairLocationInfo(l *ChairLocation) *ChairRideCache {
	globalMu.Lock()
	mu, found := chairMu[l.ChairID]
	if !found {
		mu = &sync.RWMutex{}
		chairMu[l.ChairID] = mu
	}
	mu.Lock()
	globalMu.Unlock()
	defer mu.Unlock()

	rideInfo, found := chairRideCacheMap[l.ChairID]

	if found {
		lastL := rideInfo.ChairLocation

		dis1 := math.Abs(float64(l.Latitude) - float64(lastL.Latitude))
		dis2 := math.Abs(float64(l.Longitude) - float64(lastL.Longitude))

		rideInfo.Distance = rideInfo.Distance + int64(dis1) + int64(dis2)
		rideInfo.ChairLocation = l
	} else {
		crc := &ChairRideCache{
			ChairLocation: l,
			Distance: 0,
		}
		
		chairRideCacheMap[l.ChairID] = crc
	}

	return chairRideCacheMap[l.ChairID]
}

func GetCacheChairLocationInfo(chairId string) (rideInfo *ChairRideCache, found bool) {
	globalMu.Lock()
	mu, found := chairMu[chairId]
	if !found {
		mu = &sync.RWMutex{}
		chairMu[chairId] = mu
	}
	mu.RLock()
	globalMu.Unlock()
	defer mu.RUnlock()

	rideInfo, found = chairRideCacheMap[chairId]
	return
}

func InitCacheLocationInfo(l *ChairLocation, distance int) {
	rideInfo := CacheChairLocationInfo(l)
	rideInfo.Distance = int64(distance)
}