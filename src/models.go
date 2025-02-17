package main

import (
    "encoding/json"
    "fmt"
    "net/http"

    "github.com/gorilla/mux"
    _ "github.com/lib/pq"
)

// struct to represent hubspot fields
type HubSpotField struct {
    ID        int    `json:"id"`
    CompanyID string `json:"company_id"`
    Name      string `json:"name"`
    Type      string `json:"type"`
    Value     string `json:"value"`
}

// struct to represent
type UpsoEmailCadence struct {
    ID         int    `json:"id"`
    CompanyID  string `json:"company_id"`
    CadenceID  string `json:"cadence_id"`
    Template   string `json:"template"`
    DelayHours int    `json:"delay_hours"`
}

type Config struct {
    HubSpotFields []HubSpotField     `json:"hubspot_fields"`
    UpsoCadences  []UpsoEmailCadence `json:"upso_cadences"`
}

// updates field names in PSQL instance, and invalidates Redis cache to maintain consistency.
func modifyHubSpotField(w http.ResponseWriter, r *http.Request) {
    mu.Lock() // prevent race conditions by locking
    defer mu.Unlock()

    // extract vars from request URL
    vars := mux.Vars(r)
    companyID := vars["company_id"]
    oldName := vars["old_name"]

    var update struct {
        NewName string `json:"new_name"`
    }

    // decode JSON request, populate update struct with content.
    if err := json.NewDecoder(r.Body).Decode(&update); err != nil {
        http.Error(w, err.Error(), http.StatusBadRequest)
        return
    }

    // execute UPDATE query toward DB instance, needs santizing, and properly scoping to prevent injection.
    _, err := db.Exec("UPDATE hubspot_deals SET field_name=$1 WHERE company_id=$2 AND field_name=$3", update.NewName, companyID, oldName)
    if err != nil {
        http.Error(w, "Failed to update field", http.StatusInternalServerError)
        return
    }

    // invalidate the redis cache, to force a fresh retrieval from the DB
    cacheKey := fmt.Sprintf("config:%s", companyID)
    redisClient.Del(ctx, cacheKey) // Invalidate cache

    // success response on completion &  storage.
    json.NewEncoder(w).Encode(map[string]interface{}{"message": "Field modified successfully"})
}

// inserts new fields into PSQL instance, and caches fields in redis.
func addHubSpotField(w http.ResponseWriter, r *http.Request) {
    mu.Lock()
    defer mu.Unlock()

    var field HubSpotField
    if err := json.NewDecoder(r.Body).Decode(&field); err != nil {
        http.Error(w, err.Error(), http.StatusBadRequest)
        return
    }

    _, err := db.Exec("INSERT INTO hubspot_deals (company_id, field_name, field_type, field_value) VALUES ($1, $2, $3, $4)",
        field.CompanyID, field.Name, field.Type, field.Value)
    if err != nil {
        http.Error(w, "Database insertion failed", http.StatusInternalServerError)
        return
    }

    // cache the fields in Redis, for faster response in repeat calls.
    cacheKey := fmt.Sprintf("config:%s", field.CompanyID)
    redisClient.Set(ctx, cacheKey, field, 0)

    json.NewEncoder(w).Encode(map[string]interface{}{"message": "Field added successfully", "data": field})
}

func modifyUpsoEmailCadence(w http.ResponseWriter, r *http.Request) {
    vars := mux.Vars(r)
    companyID := vars["company_id"]
    var update UpsoEmailCadence
    if err := json.NewDecoder(r.Body).Decode(&update); err != nil {
        http.Error(w, err.Error(), http.StatusBadRequest)
        return
    }
    _, err := db.Exec("UPDATE upso_email_cadences SET template=$1, delay_hours=$2 WHERE company_id=$3 AND cadence_id=$4", update.Template, update.DelayHours, companyID, update.CadenceID)
    if err != nil {
        http.Error(w, "Failed to update email cadence", http.StatusInternalServerError)
        return
    }
    fmt.Fprintf(w, "Cadence updated successfully")
}
