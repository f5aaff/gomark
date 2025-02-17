package main

import (
    "context"
    "database/sql"
    "net/http"
    "sync"

    "github.com/go-redis/redis/v8"
    "github.com/gorilla/mux"
    _ "github.com/lib/pq"
)

var (
    db          *sql.DB
    redisClient *redis.Client
    ctx         = context.Background()
    mu          sync.Mutex
)

func main() {
    initDB()
    initRedis()
    r := mux.NewRouter()
    r.HandleFunc("/hubspot/fields/add", addHubSpotField).Methods("POST")
    r.HandleFunc("/hubspot/fields/modify/{company_id}/{old_name}", modifyHubSpotField).Methods("PUT")
    r.HandleFunc("/upso/cadence/modify/{company_id}", modifyUpsoEmailCadence).Methods("PUT")
    http.ListenAndServe(":8080", r)
}
