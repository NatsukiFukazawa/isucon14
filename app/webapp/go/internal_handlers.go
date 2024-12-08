package main

import (
	"database/sql"
	"errors"
	"fmt"
	"math/rand"
	"net/http"
	"time"
)

// このAPIをインスタンス内から一定間隔で叩かせることで、椅子とライドをマッチングさせる
func internalGetMatching(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	// MEMO: 一旦最も待たせているリクエストに適当な空いている椅子マッチさせる実装とする。おそらくもっといい方法があるはず…
	ride := &Ride{}
	if err := db.GetContext(ctx, ride, `SELECT * FROM rides WHERE chair_id IS NULL ORDER BY created_at LIMIT 1`); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		writeError(w, http.StatusInternalServerError, err)
		return
	}

	matched := &Chair{}
	empty := false
	cacheChairKeys := make([]string, 0, len(chairCache))
	for key := range chairCache {
		cacheChairKeys = append(cacheChairKeys, key)
	}
	fmt.Println("chairCacheLength", len(chairCache))

	for i := 0; i < 10; i++ {
		rand.Seed(time.Now().UnixNano())
		randomKey := cacheChairKeys[rand.Intn(len(cacheChairKeys))]
		matched, found := chairCache[randomKey]
		if !found {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		if err := db.GetContext(ctx, &empty, "SELECT COUNT(*) = 0 FROM (SELECT COUNT(chair_sent_at) = 6 AS completed FROM ride_statuses WHERE ride_id IN (SELECT id FROM rides WHERE chair_id = ?) GROUP BY ride_id) is_completed WHERE completed = FALSE", matched.ID); err != nil {
			writeError(w, http.StatusInternalServerError, err)
			return
		}
		fmt.Println("matched", matched.ID, "empty", empty, "ride", ride.ID)
		if empty {
			break
		}
	}
	if !empty {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	if _, err := db.ExecContext(ctx, "UPDATE rides SET chair_id = ? WHERE id = ?", matched.ID, ride.ID); err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
