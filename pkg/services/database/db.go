package database

import (
	"database/sql"
	"eclipse/internal/logger"
	_ "github.com/mattn/go-sqlite3"
	"time"
)

func InitDB() (*sql.DB, error) {
	db, err := sql.Open("sqlite3", "../data/modules.db")
	if err != nil {
		return nil, err
	}

	_, err = db.Exec(`
        CREATE TABLE IF NOT EXISTS modules (
            id INTEGER PRIMARY KEY AUTOINCREMENT,
            wallet_address TEXT NOT NULL,
            dex_name TEXT NOT NULL,
            amount TEXT NOT NULL,
            token_name TEXT NOT NULL,
            tx_hash TEXT NOT NULL,
            created_at DATETIME DEFAULT (datetime('now', 'utc'))
        )
    `)
	if err != nil {
		return nil, err
	}

	return db, nil
}

func AddModule(db *sql.DB, wallet, dex, amount, token, txHash string) error {
	localTime := time.Now()

	logger.Info("Adding module in db: wallet=%s, dex=%s, amount=%s, token=%s, tx=%s",
		wallet, dex, amount, token, txHash)

	_, err := db.Exec(`
        INSERT INTO modules (
            wallet_address, 
            dex_name, 
            amount, 
            token_name, 
            tx_hash,
            created_at
        ) 
        VALUES (?, ?, ?, ?, ?, ?)
    `, wallet, dex, amount, token, txHash, localTime.Format("2006-01-02 15:04:05"))

	if err != nil {
		logger.Error("Error adding module: %v", err)
		return err
	}

	return nil
}
