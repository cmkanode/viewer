package dbutils

import (
	"database/sql"
)

func EnsureSchema(db *sql.DB) error {
	_, err := db.Exec(`CREATE TABLE IF NOT EXISTS slides (
		id INTEGER PRIMARY KEY,
		name TEXT,
		image BLOB,
		completed BOOLEAN NOT NULL DEFAULT FALSE
	)`)
	if err != nil {
		return err
	}
	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS tags (
		id INTEGER PRIMARY KEY,
		name TEXT NOT NULL UNIQUE
	)`)
	if err != nil {
		return err
	}
	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS slide_tags (
		id INTEGER PRIMARY KEY,
		slide_id INTEGER NOT NULL,
		tag_id INTEGER NOT NULL,
		FOREIGN KEY (slide_id) REFERENCES slides(id) ON DELETE CASCADE,
		FOREIGN KEY (tag_id) REFERENCES tags(id) ON DELETE CASCADE,
		UNIQUE(slide_id, tag_id)
	)`)
	return err
}
