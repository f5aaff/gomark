package main

import (
    "database/sql"
    "log"

    "github.com/go-redis/redis/v8"
    _ "github.com/lib/pq"
)

func initDB() {
    var err error
    db, err = sql.Open("postgres", "postgres://user:password@db-service:5432/gomark?sslmode=disable")
    if err != nil {
        log.Fatalf("Failed to connect to database: %v", err)
    }

    _, err = db.Exec(`
        CREATE TABLE IF NOT EXISTS hubspot_deals (
        id SERIAL PRIMARY KEY,
        company_id VARCHAR(255) NOT NULL,
        field_name VARCHAR(255) NOT NULL,
        field_type VARCHAR(50) NOT NULL,
        field_value TEXT NOT NULL,
        created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
    );

    CREATE TABLE IF NOT EXISTS upso_email_cadences (
        id SERIAL PRIMARY KEY,
        company_id VARCHAR(255) NOT NULL,
        cadence_id VARCHAR(255) UNIQUE NOT NULL,
        template TEXT NOT NULL,
        delay_hours INT NOT NULL,
        modified_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
    );

    -- Index to optimize lookup performance
    CREATE INDEX idx_company_id ON hubspot_fields (company_id);

    -- Ensure field names are unique per company
    CREATE UNIQUE INDEX unique_field_per_company ON hubspot_fields (company_id, name);

    -- Automatically update 'updated_at' when row changes
    CREATE OR REPLACE FUNCTION update_modified_column()
    RETURNS TRIGGER AS $$
    BEGIN
        NEW.updated_at = NOW();
        RETURN NEW;
    END;
    $$ LANGUAGE plpgsql;

    CREATE TRIGGER trigger_set_timestamp
    BEFORE UPDATE ON hubspot_fields
    FOR EACH ROW
    EXECUTE FUNCTION update_modified_column();

    `)

    if err != nil {
        log.Fatalf("database could not be instanciated:%s", err)
    }
}

func initRedis() {
    redisClient = redis.NewClient(&redis.Options{
        Addr: "redis-service:6379",
        DB:   0,
    })
}
