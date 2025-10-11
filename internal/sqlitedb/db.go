package sqlitedb

import (
	"database/sql"
	"log"

	"github.com/rogierlommers/home/internal/config"
	"github.com/sirupsen/logrus"
	_ "modernc.org/sqlite"
)

type DB struct {
	Conn *sql.DB
}

func InitDatabase(cfg config.AppConfig) *DB {
	db, err := sql.Open("sqlite", cfg.Database)
	if err != nil {
		logrus.Fatalf("failed to open db: %v", err)
	}

	_, err = db.Exec(`
        CREATE TABLE IF NOT EXISTS entry_stats (
            id INTEGER PRIMARY KEY AUTOINCREMENT,
            source TEXT NOT NULL UNIQUE,
            count INTEGER NOT NULL DEFAULT 0
        );

        CREATE TABLE IF NOT EXISTS bookmark_items (
            id INTEGER PRIMARY KEY AUTOINCREMENT,
            type TEXT,
            title TEXT,
            arg TEXT UNIQUE,
            category_id INTEGER not null,
            hide_in_gui BOOLEAN DEFAULT 0,
            FOREIGN KEY (category_id) REFERENCES bookmark_categories(id)
        );

        CREATE TABLE IF NOT EXISTS bookmark_categories (
            id INTEGER PRIMARY KEY AUTOINCREMENT,
            name TEXT UNIQUE NOT NULL,
			hide_in_gui BOOLEAN DEFAULT 0
        );
    `)
	if err != nil {
		log.Fatalf("failed to create table: %v", err)
	}

	logrus.Debugf("Database initialized, file: %s", cfg.Database)

	insertTemplateData(db)

	return &DB{
		Conn: db,
	}
}

func insertTemplateData(db *sql.DB) {

	categories := map[string]bool{
		"Personal":     false,
		"Home network": false,
		"Fun":          false,
		"Work":         false,
		"Temporary":    true,
	}

	for name, hide := range categories {
		_, err := db.Exec(`INSERT OR IGNORE INTO bookmark_categories (name, hide_in_gui) VALUES (?, ?)`, name, hide)
		if err != nil {
			logrus.Errorf("failed to insert category with name %s: %v", name, err)
		}
	}

	// then add items
	items := []Item{
		// {Type: "default", Title: "Google", Arg: "https://google.com", CategoryID: 1, HideInGUI: false},
		// {Type: "default", Title: "GitHub", Arg: "https://github.com", CategoryID: 2, HideInGUI: false},
		// {Type: "default", Title: "Personal Blog", Arg: "https://myblog.com", CategoryID: 3, HideInGUI: false},
		// {Type: "default", Title: "poep", Arg: "https://myblosg.comd", CategoryID: 4, HideInGUI: true},
		// {Type: "default", Title: "poep2", Arg: "https://mybdfsfsg.comd", CategoryID: 4, HideInGUI: true},
	}

	for _, item := range items {
		_, err := db.Exec(`INSERT OR IGNORE INTO bookmark_items (type, title, arg, autocomplete, category_id, hide_in_gui) VALUES (?, ?, ?, ?, ?, ?)`,
			item.Type, item.Title, item.Arg, item.CategoryID, item.HideInGUI)
		if err != nil {
			logrus.Errorf("failed to insert bookmark item %s: %v", item.Title, err)
		}
		logrus.Debugf("Inserted bookmark item: %+v", item)
	}
}

func (s *DB) Close() {
	if err := s.Conn.Close(); err != nil {
		logrus.Errorf("failed to close stats db: %v", err)
	}
}
