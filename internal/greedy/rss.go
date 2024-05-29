package greedy

import (
	"encoding/json"
	"strconv"
	"time"

	"github.com/boltdb/bolt"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/feeds"
	"github.com/sirupsen/logrus"
)

func (g Greedy) DisplayRSS(ginCTX *gin.Context) {
	now := time.Now()
	feed := &feeds.Feed{
		Title:       "Quick-note / personal feed",
		Link:        &feeds.Link{},
		Description: "Saved pages, all in one RSS feed",
		Created:     now,
	}

	g.db.View(func(tx *bolt.Tx) error {
		c := tx.Bucket([]byte(bucketName)).Cursor()
		count := 0
		for k, v := c.Last(); k != nil; k, v = c.Prev() {
			if count >= numberInRSS {
				break
			}

			var a *Article
			err := json.Unmarshal(v, &a)
			if err != nil {
				logrus.Error(err)
				continue
			}

			newItem := feeds.Item{
				Title:   a.Title,
				Link:    &feeds.Link{Href: a.URL},
				Created: a.Added,
				Id:      strconv.Itoa(a.ID),
			}
			feed.Add(&newItem)
			count++
		}

		return nil
	})

	rss, err := feed.ToAtom()
	if err != nil {
		logrus.Errorf("error while generating RSS feed: %s", err)
		return
	}

	ginCTX.Writer.Write([]byte(rss))
}
