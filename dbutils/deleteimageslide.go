package dbutils

import (
	"database/sql"
)

func DeleteImageSlide(db *sql.DB, slideID int) error {
	_, err := db.Exec("DELETE FROM slides WHERE id = ?", slideID)
	return err
}
