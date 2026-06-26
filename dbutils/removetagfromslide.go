package dbutils

import "database/sql"

func RemoveTagFromSlide(db *sql.DB, slideID int, tagName string) error {
	_, err := db.Exec(`DELETE FROM slide_tags 
		WHERE slide_id = ? AND tag_id = (SELECT id FROM tags WHERE name = ?)`, slideID, tagName)
	return err
}
