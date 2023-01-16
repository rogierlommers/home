package greedy

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/badoux/goscraper"
	"github.com/boltdb/bolt"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

func (g Greedy) AddArticle(c *gin.Context) {

	queryParam := c.Request.FormValue("url")
	if len(queryParam) == 0 || queryParam == "about:blank" {
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"error": "unable to insert empty or about:blank page"})
		return
	}

	newArticle := Article{
		URL:   queryParam,
		Added: time.Now(),
	}

	err := g.db.Update(func(tx *bolt.Tx) error {
		articles := tx.Bucket([]byte(bucketName))

		// Generate ID for the article.
		id, _ := articles.NextSequence()
		logrus.Infof("new sequence article: %d", id)
		newArticle.ID = int(id)

		// scrape
		err := newArticle.Scrape()
		if err != nil {
			logrus.Errorf("scraping error: %s", err)
		}

		enc, err := newArticle.encode()
		if err != nil {
			return fmt.Errorf("could not encode article: %s", err)
		}

		err = articles.Put(itob(newArticle.ID), enc)
		return err
	})

	if err != nil {
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"error": err})
		return
	}

	c.IndentedJSON(http.StatusOK, gin.H{"Message": fmt.Sprintf("Host %s added", getHostnameFromUrl(queryParam))})
}

func getHostnameFromUrl(addedUrl string) (hostname string) {
	u, err := url.Parse(addedUrl)
	if err != nil {
		logrus.Errorf("error looking up hostname [url: %s] [err: %s]", addedUrl, err)
	}
	return u.Host
}

// Scrape gathers information about new article
func (a *Article) Scrape() error {

	start := time.Now()
	logrus.Infof("start scraping article [id: %d] [url: %s]", a.ID, a.URL)

	// try to get screenshot
	var thumbnail string
	base64Image, err, debugCollection := createScreenshot(a.URL)
	if err != nil {
		logrus.Error(err)
	} else {
		thumbnail = base64Image
	}

	// scrape html
	s, err := goscraper.Scrape(a.URL, 5)
	if err != nil {
		a.Title = fmt.Sprintf("[Greedy] scrape failed: %q", a.URL)
		a.Description = fmt.Sprintf("Scraping failed for url %q", a.URL)
		logrus.Errorf("scrape error: %s", err)
	} else {
		a.Title = fmt.Sprintf("[Greedy] %s", s.Preview.Title)
		a.Description = fmt.Sprintf("%s<img src=\"data:image/jpeg;base64, %s\"/>", s.Preview.Description, thumbnail)

		// now add devug info to description field
		for _, v := range debugCollection {
			a.Description = a.Description + "<br/>" + v
		}

	}

	// debugging info
	elapsed := time.Since(start)
	logrus.Infof("scraping done, id: %d, title: %q, elapsed: %s, debugs: %d", a.ID, a.Title, elapsed, len(debugCollection))
	return nil
}

// itob returns an 8-byte big endian representation of v.
func itob(v int) []byte {
	b := make([]byte, 8)
	binary.BigEndian.PutUint64(b, uint64(v))
	return b
}

func (a *Article) encode() ([]byte, error) {
	enc, err := json.Marshal(a)
	if err != nil {
		return nil, err
	}
	return enc, nil
}
