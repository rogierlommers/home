package sqlitedb

import (
	"crypto/sha256"
	"encoding/hex"

	"github.com/sirupsen/logrus"
)

type Bookmarks struct {
	Cache struct {
		Seconds int `json:"seconds"`
	} `json:"cache"`
	Items []Item `json:"items"`
}

type Item struct {
	UID        string `json:"uid"` // Alfred specific, used for sorting
	ID         int    `json:"id"`  // id in the database
	Type       string `json:"type"`
	Title      string `json:"title"`
	Arg        string `json:"arg"`
	CategoryID int    `json:"category_id"`
	HideInGUI  bool   `json:"hide_in_gui,omitempty"`
}

type Category struct {
	ID        int    `json:"id"`
	Name      string `json:"name"`
	HideInGUI bool   `json:"hide_in_gui,omitempty"`
}

func (s *DB) GetBookmarks() (Bookmarks, error) {
	var Bookmarks Bookmarks
	Bookmarks.Cache.Seconds = 3600 // tell Alfred to cache for 1 hour

	rows, err := s.db.Query(`
    SELECT b.id, b.type, b.title, b.arg, b.category_id, b.hide_in_gui
    FROM bookmark_items b
    ORDER BY b.id ASC
`)
	if err != nil {
		return Bookmarks, err
	}
	defer rows.Close()

	for rows.Next() {
		var i Item

		if err := rows.Scan(&i.ID, &i.Type, &i.Title, &i.Arg, &i.CategoryID, &i.HideInGUI); err != nil {
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

func (s *DB) GetCategories(excludeHidden bool) ([]Category, error) {
	var (
		categories []Category
		query      string
	)

	if excludeHidden {
		logrus.Debugf("Excluding hidden categories from results")
		query = "SELECT id, name, hide_in_gui FROM bookmark_categories WHERE hide_in_gui != true ORDER BY id ASC"
	} else {
		logrus.Debugf("Including all categories in results")
		query = "SELECT id, name, hide_in_gui FROM bookmark_categories ORDER BY id ASC"
	}

	rows, err := s.db.Query(query)
	if err != nil {
		return categories, err
	}
	defer rows.Close()

	for rows.Next() {
		var c Category
		if err := rows.Scan(&c.ID, &c.Name, &c.HideInGUI); err != nil {
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
		INSERT INTO bookmark_items (type, title, arg, category_id, hide_in_gui)
		VALUES (?, ?, ?, ?, ?)`, item.Type, item.Title, item.Arg, item.CategoryID, item.HideInGUI)
	return err
}

func convertSHA256(input string) string {
	// Generate SHA256 hash
	hash := sha256.Sum256([]byte(input))

	// Convert to hexadecimal string
	return hex.EncodeToString(hash[:])
}
