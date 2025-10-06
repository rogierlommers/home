package sqlitedb

import (
	"database/sql"
	"log"

	"github.com/rogierlommers/home/internal/config"
	"github.com/sirupsen/logrus"
	_ "modernc.org/sqlite"
)

type DB struct {
	db *sql.DB
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
            autocomplete TEXT,
            category_id INTEGER not null,
            FOREIGN KEY (category_id) REFERENCES bookmark_categories(id)
        );

        CREATE TABLE IF NOT EXISTS bookmark_categories (
            id INTEGER PRIMARY KEY AUTOINCREMENT,
            name TEXT UNIQUE NOT NULL
        );
    `)
	if err != nil {
		log.Fatalf("failed to create table: %v", err)
	}

	insertTemplateData(db)

	return &DB{
		db: db,
	}
}

func insertTemplateData(db *sql.DB) {

	// first add categories
	categories := []string{"Personal", "Home network", "Fun", "Work"}
	for _, cat := range categories {
		_, err := db.Exec(`INSERT OR IGNORE INTO bookmark_categories (name) VALUES (?)`, cat)
		if err != nil {
			logrus.Errorf("failed to insert category %s: %v", cat, err)
		}
	}

	// then add items
	items := []Item{
		// {Type: "default", Title: "Google", Arg: "https://google.com", Autocomplete: "google", CategoryID: 1},
		// {Type: "default", Title: "GitHub", Arg: "https://github.com", Autocomplete: "github", CategoryID: 2},
		// {Type: "default", Title: "Personal Blog", Arg: "https://myblog.com", Autocomplete: "blog", CategoryID: 3},
		// {Type: "default", Title: "poep", Arg: "https://myblosg.comd", Autocomplete: "blog", CategoryID: 4},
		// {Type: "default", Title: "poep2", Arg: "https://mybdfsfsg.comd", Autocomplete: "bloghaha", CategoryID: 4},
	}

	for _, item := range items {
		_, err := db.Exec(`INSERT OR IGNORE INTO bookmark_items (type, title, arg, autocomplete, category_id) VALUES (?, ?, ?, ?, ?)`,
			item.Type, item.Title, item.Arg, item.Autocomplete, item.CategoryID)
		if err != nil {
			logrus.Errorf("failed to insert bookmark item %s: %v", item.Title, err)
		}
		logrus.Debugf("Inserted bookmark item: %+v", item)
	}

}

func (s *DB) Close() {
	if err := s.db.Close(); err != nil {
		logrus.Errorf("failed to close stats db: %v", err)
	}
}
