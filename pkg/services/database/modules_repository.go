package database

import (
	"database/sql"
	"eclipse/internal/logger"
)

func GetModuleCountForWallet(db *sql.DB, walletAddress, moduleName string) (int, error) {
	var count int

	err := db.QueryRow(`
        SELECT COUNT(*) 
        FROM modules 
        WHERE wallet_address = ? AND dex_name = ?
    `, walletAddress, moduleName).Scan(&count)

	if err != nil {
		logger.Error("Error getting module count for wallet %s and module %s: %v",
			walletAddress, moduleName, err)
		return 0, err
	}

	return count, nil
}

func GetAllModuleCountsForWallet(db *sql.DB, walletAddress string) (map[string]int, error) {
	counts := make(map[string]int)

	rows, err := db.Query(`
        SELECT dex_name, COUNT(*) as count 
        FROM modules 
        WHERE wallet_address = ? 
        GROUP BY dex_name
    `, walletAddress)

	if err != nil {
		logger.Error("Error getting all module counts for wallet %s: %v",
			walletAddress, err)
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var moduleName string
		var count int

		if err := rows.Scan(&moduleName, &count); err != nil {
			logger.Error("Error scanning row: %v", err)
			return nil, err
		}

		counts[moduleName] = count
	}

	if err = rows.Err(); err != nil {
		logger.Error("Error iterating rows: %v", err)
		return nil, err
	}

	return counts, nil
}
