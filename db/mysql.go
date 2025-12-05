package db

import (
    "database/sql"
    _ "github.com/go-sql-driver/mysql"
)

func OpenFromDSN(dsn string) (*sql.DB, error){
    db, err := sql.Open("mysql", dsn)
    if err!=nil { return nil, err }
    if err := db.Ping(); err!=nil { return nil, err }
    return db, nil
}

func Migrate(db *sql.DB) error{
    stmts := []string{
        `CREATE TABLE IF NOT EXISTS customers (
            id BIGINT AUTO_INCREMENT PRIMARY KEY,
            phone VARCHAR(32) NOT NULL UNIQUE,
            first_name VARCHAR(100),
            last_name VARCHAR(100),
            metadata TEXT
        );`,
        `CREATE TABLE IF NOT EXISTS campaigns (
            id BIGINT AUTO_INCREMENT PRIMARY KEY,
            name VARCHAR(200) NOT NULL,
            template TEXT NOT NULL,
            status VARCHAR(50) NOT NULL DEFAULT 'draft',
            created_at DATETIME DEFAULT CURRENT_TIMESTAMP
        );`,
        `CREATE TABLE IF NOT EXISTS outbound_messages (
            id BIGINT AUTO_INCREMENT PRIMARY KEY,
            campaign_id BIGINT NOT NULL,
            customer_id BIGINT NOT NULL,
            to_phone VARCHAR(32) NOT NULL,
            body TEXT NOT NULL,
            status VARCHAR(50) NOT NULL DEFAULT 'queued',
            created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
            FOREIGN KEY (campaign_id) REFERENCES campaigns(id) ON DELETE CASCADE,
            FOREIGN KEY (customer_id) REFERENCES customers(id) ON DELETE CASCADE
        );`,
    }
    for _, s := range stmts {
        if _, err := db.Exec(s); err!=nil { return err }
    }
    return nil
}
