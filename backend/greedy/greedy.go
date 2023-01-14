package greedy

import (
	"fmt"
	"time"

	"github.com/boltdb/bolt"
	cfg "github.com/rogierlommers/quick-note/backend/config"
	"github.com/sirupsen/logrus"
)

type Greedy struct {
	db   *bolt.DB
	open bool
}

const (
	bucketName       = "articles"
	keep             = 100   // amount of records to keep on disk
	numberInRSS      = 100   // amount of records to display in feed
	cleanupFrequency = 86400 // 1 day
)

// Article holds information about saved URLs
type Article struct {
	ID          int
	URL         string
	Title       string
	Description string
	Added       time.Time
}

func NewGreedy() (Greedy, error) {
	boltConfig := &bolt.Options{Timeout: 1 * time.Second}

	db, err := bolt.Open(cfg.Settings.GreedyFile, 0600, boltConfig)
	if err != nil {
		return Greedy{}, err
	}

	logrus.Infof("file %s created or opened", cfg.Settings.GreedyFile)

	db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists([]byte(bucketName))
		if err != nil {
			return fmt.Errorf("create bucket: %s", err)
		}
		return nil
	})

	return Greedy{
		db:   db,
		open: true,
	}, nil
}

func (g *Greedy) CloseArticleDB() {
	logrus.Info("closing greedy database file")
	if err := g.db.Close(); err != nil {
		logrus.Fatal(err)
	}

	g.open = false
}

func (g Greedy) Count() (amount int) {
	g.db.View(func(tx *bolt.Tx) error {
		c := tx.Bucket([]byte(bucketName)).Cursor()
		amount = c.Bucket().Stats().KeyN
		return nil
	})

	return amount
}
