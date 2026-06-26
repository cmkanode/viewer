package dbutils

import (
	"database/sql"
)

func CountSlides(db *sql.DB) (int, error) {
	row := db.QueryRow("SELECT COUNT(*) FROM slides")
	var count int
	if err := row.Scan(&count); err != nil {
		return 0, err
	}
	return count, nil
}
