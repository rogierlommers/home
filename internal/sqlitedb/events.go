package sqlitedb

import "github.com/sirupsen/logrus"

func (s *DB) GetEventsCategories() ([]string, error) {
	var eventCategories []string
	query := "SELECT DISTINCT category AS category_name FROM events WHERE category IS NOT NULL AND category != ''"

	rows, err := s.Conn.Query(query)
	if err != nil {
		logrus.Errorf("Failed to query event categories: %v", err)
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var c string
		if err := rows.Scan(&c); err != nil {
			logrus.Errorf("Failed to scan category row: %v", err)
			return nil, err
		}
		eventCategories = append(eventCategories, c)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return eventCategories, nil
}

func (s *DB) DeleteOldEvents() (int, error) {
	// delete all events older than 30 days
	result, err := s.Conn.Exec("DELETE FROM events WHERE added < datetime('now', '-30 days')")
	if err != nil {
		logrus.Errorf("Failed to delete old events: %v", err)
		return 0, err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		logrus.Errorf("Failed to get rows affected after deleting old events: %v", err)
		return 0, err
	}

	return int(rowsAffected), nil
}
