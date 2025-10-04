package stats

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

type EntryCount struct {
	Source string `json:"source"`
	Count  int    `json:"count"`
}

func InitStatsDB(cfg config.AppConfig) *DB {
	db, err := sql.Open("sqlite", cfg.StatsDB)
	if err != nil {
		logrus.Fatalf("failed to open db: %v", err)
	}

	_, err = db.Exec(`
        CREATE TABLE IF NOT EXISTS entry_stats (
            id INTEGER PRIMARY KEY AUTOINCREMENT,
            source TEXT NOT NULL UNIQUE,
            count INTEGER NOT NULL DEFAULT 0
        )
    `)
	if err != nil {
		log.Fatalf("failed to create table: %v", err)
	}

	return &DB{
		db: db,
	}
}

func (s *DB) IncrementEntry(source string) error {
	res, err := s.db.Exec(`UPDATE entry_stats SET count = count + 1 WHERE source = ?`, source)
	if err != nil {
		return err
	}
	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		_, err = s.db.Exec(`INSERT INTO entry_stats (source, count) VALUES (?, 1)`, source)
		if err != nil {
			return err
		}
		logrus.Debugf("created entry for new source %s", source)
	} else {
		logrus.Debugf("incremented entry for source %s", source)
	}

	return nil
}

func (s *DB) GetEntryCount(source string) (int, error) {
	var count int
	err := s.db.QueryRow(`SELECT count FROM entry_stats WHERE source = ?`, source).Scan(&count)
	return count, err
}

func (s *DB) GetAllEntryCounts() ([]EntryCount, error) {
	rows, err := s.db.Query(`SELECT source, count FROM entry_stats ORDER BY source ASC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var counts []EntryCount
	for rows.Next() {
		var ec EntryCount
		if err := rows.Scan(&ec.Source, &ec.Count); err != nil {
			return nil, err
		}
		counts = append(counts, ec)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return counts, nil
}

func (s *DB) Close() {
	if err := s.db.Close(); err != nil {
		logrus.Errorf("failed to close stats db: %v", err)
	}
}
