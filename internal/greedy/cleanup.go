package greedy

import (
	"time"

	"github.com/sirupsen/logrus"
)

func (g *Greedy) scheduleCleanup() {

	// Schedule cleanup every cleanupFrequency seconds
	go func() {
		logrus.Infof("starting scheduled cleanup every %d seconds", g.cfg.GreedyCleanupFrequency)

		for {
			time.Sleep(time.Duration(g.cfg.GreedyCleanupFrequency) * time.Second)

			if err := g.deleteOldRecords(keep); err != nil {
				logrus.Errorf("error during cleanup of old records: %v", err)
			}

			logrus.Debugf("completed scheduled scraping of articles")
		}
	}()

}

func (g *Greedy) deleteOldRecords(keep int) error {
	query := `DELETE FROM greedy_urls
			  WHERE id NOT IN (
				  SELECT id FROM greedy_urls
				  ORDER BY id DESC
				  LIMIT ?
			  );`
	_, err := g.db.Conn.Exec(query, keep)
	return err
}
