package dbutils

import (
	"database/sql"

	. "viewer/models"
)

func SelectRandomSlide(db *sql.DB) (*Slide, error) {
	row := db.QueryRow("SELECT id, name, image, completed FROM slides WHERE completed = FALSE ORDER BY RANDOM() LIMIT 1")
	var slide Slide
	if err := row.Scan(&slide.ID, &slide.Name, &slide.Image, &slide.Completed); err != nil {
		return nil, err
	}
	tags, err := GetTagsForSlide(db, slide.ID)
	if err != nil {
		return nil, err
	}
	slide.Tags = tags
	return &slide, nil
}
