package greedy

import (
	"bytes"
	"encoding/base64"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/badoux/goscraper"
	"github.com/boltdb/bolt"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

func (g Greedy) AcceptedResponse(c *gin.Context) {

	encodedMessage := c.Query("msg")

	var newArticle Article
	if err := decodeFromBase64(&newArticle, encodedMessage); err != nil {
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// build html
	output := fmt.Sprintf(`<!DOCTYPE html>
	<html lang="en">
	
	<head>
	  <meta charset="utf-8" />
	  <meta name="viewport" content="width=device-width, initial-scale=1" />
	  <title>quick-note | url added</title>
	  <link rel="stylesheet" href="https://fonts.googleapis.com/css?family=Roboto:300,300italic,700,700italic" />
	  <link rel="stylesheet" href="https://cdnjs.cloudflare.com/ajax/libs/normalize/8.0.1/normalize.css" />
	  <link rel="stylesheet" href="https://cdnjs.cloudflare.com/ajax/libs/milligram/1.4.1/milligram.min.css" />
	  <link rel="stylesheet" href="https://milligram.io/styles/main.css" />
	</head>
	
	<body>
	  <main class="wrapper">
	
		<section class="container" id="examples">
		  <h5 class="title">Success!</h5>
		  <p>the url has succesfully been added</p>
		  <p><strong>Description:</strong><br/>%s<br/></p>
		  <p><strong>Title:</strong><br/>%s<br/></p>	  
		</section>
	
	  </main>
	
	</body>
	
	</html>`, newArticle.Description, newArticle.Title)

	// serve
	c.Header("Content-Type", "text/html")
	c.String(200, output)
}

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

	// base64 encode the message
	msg, err := encodeToBase64(newArticle)
	if err != nil {
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// redirect with encoded message
	c.Redirect(http.StatusTemporaryRedirect, fmt.Sprintf("/api/greedy/accepted?msg=%s", msg))
}

// Scrape gathers information about new article
func (a *Article) Scrape() error {

	start := time.Now()
	logrus.Infof("start scraping article [id: %d] [url: %s]", a.ID, a.URL)

	// scrape html
	s, err := goscraper.Scrape(a.URL, 5)
	if err != nil {
		a.Title = fmt.Sprintf("[Greedy] scrape failed: %q", a.URL)
		a.Description = fmt.Sprintf("Scraping failed for url %q", a.URL)
		logrus.Errorf("scrape error: %s", err)
	} else {
		a.Title = fmt.Sprintf("[Greedy] %s", s.Preview.Title)
		a.Description = s.Preview.Description
	}

	// debugging info
	elapsed := time.Since(start)
	logrus.Infof("scraping done, id: %d, title: %q, elapsed: %s", a.ID, a.Title, elapsed)
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

func encodeToBase64(v interface{}) (string, error) {
	var buf bytes.Buffer
	encoder := base64.NewEncoder(base64.StdEncoding, &buf)
	err := json.NewEncoder(encoder).Encode(v)
	if err != nil {
		return "", err
	}
	encoder.Close()
	return buf.String(), nil
}

func decodeFromBase64(v interface{}, enc string) error {
	return json.NewDecoder(base64.NewDecoder(base64.StdEncoding, strings.NewReader(enc))).Decode(v)
}
