package dbutils

import (
	"database/sql"
)

func GetTagsForSlide(db *sql.DB, slideID int) ([]string, error) {
	rows, err := db.Query(`SELECT t.name FROM tags t
		INNER JOIN slide_tags st ON t.id = st.tag_id
		WHERE st.slide_id = ?
		ORDER BY t.name`, slideID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tags []string
	for rows.Next() {
		var tag string
		if err := rows.Scan(&tag); err != nil {
			return nil, err
		}
		tags = append(tags, tag)
	}
	return tags, rows.Err()
}
