package sqlitedb

import (
	"database/sql"

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
            title TEXT,
            arg TEXT UNIQUE,
            category_id INTEGER not null,
            hide_in_gui BOOLEAN DEFAULT 0,
            priority INTEGER,
            FOREIGN KEY (category_id) REFERENCES bookmark_categories(id)
        );

        CREATE TABLE IF NOT EXISTS bookmark_categories (
            id INTEGER PRIMARY KEY AUTOINCREMENT,
            name TEXT UNIQUE NOT NULL,
            hide_in_gui BOOLEAN DEFAULT 0
        );

        CREATE TABLE IF NOT EXISTS events (
            id INTEGER PRIMARY KEY AUTOINCREMENT,
            source TEXT,
            label TEXT NOT NULL,
            message TEXT NOT NULL,
            category TEXT,
            added TIMESTAMP DEFAULT CURRENT_TIMESTAMP
        );

        CREATE INDEX IF NOT EXISTS idx_ha_events_categories
        ON events (category);

        CREATE INDEX IF NOT EXISTS idx_ha_events_labels
        ON events (label);

        `)

	if err != nil {
		logrus.Fatalf("failed to create table: %v", err)
	}

	logrus.Debugf("Database initialized, file: %s", cfg.Database)

	createCategories(db)

	return &DB{
		Conn: db,
	}
}

func createCategories(db *sql.DB) {

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

}

func (s *DB) Close() {
	if err := s.Conn.Close(); err != nil {
		logrus.Errorf("failed to close stats db: %v", err)
	}
}
