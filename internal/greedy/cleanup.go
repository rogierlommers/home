package greedy

import (
	"fmt"
	"time"

	"github.com/boltdb/bolt"
	"github.com/rogierlommers/home/internal/prom_error"
	"github.com/sirupsen/logrus"
)

func (g Greedy) ScheduleCleanup() {
	go func() {
		logrus.Infof("scheduled greedy cleanup, every %d seconds, remove more than %d records", cleanupFrequency, keep)
		for {
			deleted := g.cleanUp(keep)
			logrus.Infof("deleted %d records from greedy database", deleted)
			time.Sleep(cleanupFrequency * time.Second)
		}
	}()

}

func (g Greedy) cleanUp(numberToKeep int) int {
	var (
		count   = 0
		deleted = 0
	)

	g.db.Update(func(tx *bolt.Tx) error {
		c := tx.Bucket([]byte(bucketName)).Cursor()

		for k, _ := c.Last(); k != nil; k, _ = c.Prev() {
			count++
			if count > numberToKeep {
				err := c.Delete()
				if err != nil {
					prom_error.LogError(fmt.Sprintf("error deleting record while cleanup: %q", err))
				} else {
					deleted++
				}
			}
		}
		return nil
	})

	return deleted
}
