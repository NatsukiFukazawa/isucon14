package main

import (
	"sync"
	"time"
)

// 新規追加管理
var globalRideStatusMu = sync.RWMutex{}
// 固有操作管理 ride_id: RideStatus
var rideStatusMu = make(map[string]*sync.RWMutex)

// ride_idをkeyに最新のride_statusのみ保持する
var rideStatusCacheMap = make(map[string]*RideStatus)

// キャッシュ構造体作るの面倒な時に
func CacheRideStatusWrapper(id string, rideId string, status string, createdAt time.Time, appSentAt *time.Time, chairSentAt *time.Time) {
	rs := &RideStatus{
		ID: id,
		RideID: rideId,
		Status: status,
		CreatedAt: createdAt,
		AppSentAt: appSentAt,
		ChairSentAt: chairSentAt,
	}
	CacheRideStatus(rs)
}

// キャッシュロジック
func CacheRideStatus(rs *RideStatus) {
	rideId := rs.RideID
	globalRideStatusMu.Lock()
	mu, found := rideStatusMu[rideId]
	if !found {
		mu = &sync.RWMutex{}
		rideStatusMu[rideId] = mu
	}
	mu.Lock()
	globalRideStatusMu.Unlock()
	defer mu.Unlock()

	rideStatusCacheMap[rideId] = rs
}

func GetCacheLatestRideStatus(rideId string) (rideStatus *RideStatus, found bool) {
	globalRideStatusMu.Lock()
	mu, found := rideStatusMu[rideId]
	if !found {
		mu = &sync.RWMutex{}
		rideStatusMu[rideId] = mu
	}
	mu.RLock()
	globalRideStatusMu.Unlock()
	defer mu.RUnlock()

	rideStatus, found = rideStatusCacheMap[rideId]
	return
}

// 最新のride_statusのapp_sent_atのみ更新する
// ride_id, ride_status_idが一致した場合のみ
func UpdateCacheRideStatusAppSentAt(rideStatusId string, rideId string, appSentAt time.Time) {
	mu, _ := rideStatusMu[rideId]
	mu.Lock()
	defer mu.Unlock()

	rideStatus, found := rideStatusCacheMap[rideId]
	if found && rideStatus.ID == rideStatusId {
		rideStatus.AppSentAt = &appSentAt
	}
}

// 最新のride_statusのchair_sent_atのみ更新する
// ride_id, ride_status_idが一致した場合のみ
func UpdateCacheRideChairSentAt(rideStatusId string, rideId string, chairSentAt time.Time) {
	mu, _ := rideStatusMu[rideId]
	mu.Lock()
	defer mu.Unlock()

	rideStatus, found := rideStatusCacheMap[rideId]
	if found && rideStatus.ID == rideStatusId {
		rideStatus.AppSentAt = &chairSentAt
	}
}

func InitRideStatus(rs *RideStatus) {
	CacheRideStatus(rs)
}