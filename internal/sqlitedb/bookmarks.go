package sqlitedb

import (
	"crypto/sha256"
	"encoding/hex"

	"github.com/sirupsen/logrus"
)

type Bookmarks struct {
	Items []Item `json:"items"`
}

type Item struct {
	UID          string `json:"uid"` // Alfred specific, used for sorting
	ID           int    `json:"id"`  // id in the database
	Type         string `json:"type"`
	Title        string `json:"title"`
	Arg          string `json:"arg"`
	Autocomplete string `json:"autocomplete"`
	CategoryID   int    `json:"category_id"`
}

type Categories struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

func (s *DB) GetBookmarks() (Bookmarks, error) {
	var Bookmarks Bookmarks

	rows, err := s.db.Query(`
    SELECT b.id, b.type, b.title, b.arg, b.autocomplete, b.category_id
    FROM bookmark_items b
    ORDER BY b.id ASC
`)
	if err != nil {
		return Bookmarks, err
	}
	defer rows.Close()

	for rows.Next() {
		var i Item

		if err := rows.Scan(&i.ID, &i.Type, &i.Title, &i.Arg, &i.Autocomplete, &i.CategoryID); err != nil {
			logrus.Errorf("Failed to scan row: %v", err)
			return Bookmarks, err
		}
		i.UID = convertSHA256(i.Arg)
		Bookmarks.Items = append(Bookmarks.Items, i)
	}

	if err := rows.Err(); err != nil {
		return Bookmarks, err
	}

	return Bookmarks, nil
}

func (s *DB) GetCategories() ([]Categories, error) {
	var categories []Categories

	rows, err := s.db.Query(`SELECT id, name FROM bookmark_categories ORDER BY name ASC`)
	if err != nil {
		return categories, err
	}
	defer rows.Close()

	for rows.Next() {
		var c Categories
		if err := rows.Scan(&c.ID, &c.Name); err != nil {
			logrus.Errorf("Failed to scan category row: %v", err)
			return categories, err
		}
		categories = append(categories, c)
	}

	if err := rows.Err(); err != nil {
		return categories, err
	}

	return categories, nil
}

func (s *DB) AddBookmark(item Item) error {
	_, err := s.db.Exec(`
	INSERT INTO bookmark_items (type, title, arg, autocomplete, category_id)
	VALUES (?, ?, ?, ?, ?)
`, item.Type, item.Title, item.Arg, item.Autocomplete, item.CategoryID)
	return err
}

func convertSHA256(input string) string {
	// Generate SHA256 hash
	hash := sha256.Sum256([]byte(input))
	// Convert to hexadecimal string
	return hex.EncodeToString(hash[:])
}
