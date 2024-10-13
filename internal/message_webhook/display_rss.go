package message_webhook

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/feeds"
	"github.com/sirupsen/logrus"
)

func displayRSS(ginCTX *gin.Context) {

	now := time.Now()
	feed := &feeds.Feed{
		Title:       "Home service / pushed webhook items",
		Description: "Home service / pushed webhook items",
		Link:        &feeds.Link{},
		Created:     now,
	}

	for _, cachedItem := range cache.GetElements() {
		logrus.Infof("item: %s", cachedItem.(message).Message)
		newItem := feeds.Item{
			Title: cachedItem.(message).Message,
			// Link:    &feeds.Link{Href: a.URL},
			Created: cachedItem.(message).Timestamp,
			Id:      cachedItem.(message).ID,
		}

		feed.Add(&newItem)
	}

	rss, err := feed.ToAtom()
	if err != nil {
		logrus.Errorf("error while generating RSS feed: %s", err)
		return
	}

	ginCTX.Writer.Write([]byte(rss))
}
