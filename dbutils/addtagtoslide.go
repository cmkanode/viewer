package dbutils

import (
	"database/sql"
	"fmt"
)

func AddTagToSlide(db *sql.DB, slideID int, tagName string) error {
	if tagName == "" {
		return fmt.Errorf("tag name cannot be empty")
	}
	// Insert tag if it doesn't exist
	_, err := db.Exec("INSERT OR IGNORE INTO tags (name) VALUES (?)", tagName)
	if err != nil {
		return err
	}

	// Get the tag ID
	var tagID int
	err = db.QueryRow("SELECT id FROM tags WHERE name = ?", tagName).Scan(&tagID)
	if err != nil {
		return err
	}

	// Insert the relationship
	_, err = db.Exec("INSERT OR IGNORE INTO slide_tags (slide_id, tag_id) VALUES (?, ?)", slideID, tagID)
	return err
}
