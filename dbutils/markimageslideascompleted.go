package dbutils

import (
	"database/sql"
)

func MarkImageSlideAsCompleted(db *sql.DB, slideID int) error {
	_, err := db.Exec("UPDATE slides SET completed = TRUE WHERE id = ?", slideID)
	return err
}
