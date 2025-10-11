package greedy

import (
	"encoding/base64"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gocolly/colly"
	"github.com/gorilla/feeds"
	"github.com/rogierlommers/home/internal/config"
	"github.com/rogierlommers/home/internal/sqlitedb"
	"github.com/sirupsen/logrus"
)

const (
	keep                 = 250   // amount of records to keep
	numberInRSS          = 100   // amount of records to display in feed
	cleanupFrequency     = 86400 // 1 day
	scrapingFrequncy     = 3600  // 1 hour
	userAgentForScraping = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/88.0.4324.182 Safari/537.36"
)

type GreedyURL struct {
	ID    int
	URL   string
	Title string
	Added time.Time
}

type Greedy struct {
	db *sqlitedb.DB
}

func NewGreedy(router *gin.Engine, cfg config.AppConfig, db *sqlitedb.DB) (Greedy, error) {

	// create instance
	g := Greedy{db: db}

	// add routes
	router.GET("/api/greedy/add", g.addURL)
	router.GET("/api/greedy/rss", g.displayRSS)
	router.GET("/api/greedy/accepted", g.AcceptedResponse)

	// if not exist, create table
	if err := g.createTable(); err != nil {
		return Greedy{}, err
	}

	// schedule scraping
	g.scheduleScraping()

	// schedule cleanup
	g.scheduleCleanup()
	return g, nil
}

// AddURL adds a new url to the database.
func (g *Greedy) addURL(c *gin.Context) {
	newURL := c.Request.FormValue("url")
	if len(newURL) == 0 || newURL == "about:blank" {
		c.IndentedJSON(500, gin.H{"error": "unable to insert empty or about:blank page"})
		return
	}

	greedyURL := GreedyURL{
		URL:   newURL,
		Added: time.Now(),
	}

	if err := g.saveURL(greedyURL); err != nil {
		c.IndentedJSON(500, gin.H{"error": err.Error()})
	}

	// base64 encode the title
	encoded := base64.StdEncoding.EncodeToString([]byte(greedyURL.URL))

	// redirect with encoded message
	c.Redirect(http.StatusTemporaryRedirect, fmt.Sprintf("/api/greedy/accepted?msg=%s", encoded))
}

func (g *Greedy) displayRSS(c *gin.Context) {

	now := time.Now()

	// create the feed
	feed := &feeds.Feed{
		Title:       "Quick-note / personal feed",
		Link:        &feeds.Link{},
		Description: "Saved pages, all in one RSS feed",
		Created:     now,
	}

	// load all articles
	onlyIncludeScrapedUrls := true
	articles, err := g.getURLs(onlyIncludeScrapedUrls)
	if err != nil {
		c.IndentedJSON(500, gin.H{"error": err.Error()})
		return
	}

	// add articles to the feed
	var newItem *feeds.Item

	for _, a := range articles {
		newItem = &feeds.Item{
			Title:   a.Title,
			Link:    &feeds.Link{Href: a.URL},
			Created: a.Added,
			Id:      strconv.Itoa(a.ID),
		}

		feed.Add(newItem)
	}

	rss, err := feed.ToAtom()
	if err != nil {
		logrus.Errorf("error while generating RSS feed: %s", err)
		c.IndentedJSON(500, gin.H{"error": err.Error()})
		return
	}

	c.Writer.Write([]byte(rss))
}

func (g *Greedy) createTable() error {

	_, err := g.db.Conn.Exec(`
        CREATE TABLE IF NOT EXISTS greedy_urls (
            id INTEGER PRIMARY KEY AUTOINCREMENT,
            url TEXT NOT NULL,
            title TEXT NOT NULL,
            scrape_done BOOLEAN NOT NULL DEFAULT 0,
            date_added TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
        )
    `)

	return err
}

func (g Greedy) saveURL(greedyURL GreedyURL) error {

	_, err := g.db.Conn.Exec(`
		INSERT INTO greedy_urls (url, title, date_added)
		VALUES (?, ?, ?)
	`, greedyURL.URL, greedyURL.Title, greedyURL.Added)

	return err
}

func (g *Greedy) updateURL(greedyURL GreedyURL) error {

	// update article's title and mark as scraped
	_, err := g.db.Conn.Exec(`
		UPDATE greedy_urls
		SET title = ?, scrape_done = true
		WHERE id = ?
	`, greedyURL.Title, greedyURL.ID)
	return err
}

func (g Greedy) getURLs(onlyScrapedURLs bool) ([]GreedyURL, error) {
	var (
		urls  []GreedyURL
		query string
	)

	if onlyScrapedURLs {
		query = `SELECT id, url, title, date_added FROM greedy_urls WHERE scrape_done = true ORDER BY date_added DESC LIMIT ?`
	} else {
		query = `SELECT id, url, title, date_added FROM greedy_urls WHERE scrape_done = false ORDER BY date_added DESC LIMIT ?`
	}

	rows, err := g.db.Conn.Query(query, numberInRSS)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var url GreedyURL
		if err := rows.Scan(&url.ID, &url.URL, &url.Title, &url.Added); err != nil {
			return nil, err
		}
		urls = append(urls, url)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return urls, nil
}

func (u *GreedyURL) scrape() error {

	// init colly scraper
	c := colly.NewCollector(
		colly.MaxDepth(5),
		colly.UserAgent(userAgentForScraping),
	)

	// find and set title
	c.OnHTML("title", func(e *colly.HTMLElement) {
		if u.Title == "" {
			u.Title = e.Text
			return
		}
	})

	c.OnError(func(r *colly.Response, err error) {
		logrus.Errorf("Request URL: %s failed with error: %s", r.Request.URL.String(), err)
	})

	if err := c.Visit(u.URL); err != nil {
		return err
	}

	logrus.Debugf("scraped title: %s", u.Title)
	return nil
}

func getBaseURL(fullURL string) string {
	u, err := url.Parse(fullURL)
	if err != nil {
		return fullURL
	}

	parsedUrl := fmt.Sprintf("%s%s", u.Host, u.Path)
	return parsedUrl
}

func (g Greedy) AcceptedResponse(c *gin.Context) {
	decodedMessage, err := base64.StdEncoding.DecodeString(c.Query("msg"))
	if err != nil {
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// build html
	output := fmt.Sprintf(`<!DOCTYPE html>
	<html lang="en">
	
	<head>
	  <meta charset="utf-8" />
	  <meta name="viewport" content="width=device-width, initial-scale=1" />
	  <title>home | url added</title>
	  <link rel="stylesheet" href="https://fonts.googleapis.com/css?family=Roboto:300,300italic,700,700italic" />
	  <link rel="stylesheet" href="https://cdnjs.cloudflare.com/ajax/libs/normalize/8.0.1/normalize.css" />
	  <link rel="stylesheet" href="https://cdnjs.cloudflare.com/ajax/libs/milligram/1.4.1/milligram.min.css" />
	  <link rel="stylesheet" href="https://milligram.io/styles/main.css" />
	</head>
	
	<body>
	  <main class="wrapper">
	
		<section class="container" id="examples">
		  <h1 class="title"><a>Success!</a></h1>
		  <p><em>The url has succesfully been added.</em></p>
		  <p><strong>Title:</strong><br/>%s<br/></p>	  
		</section>

	  </main>
	
	</body>
	
	</html>`, string(decodedMessage))

	// serve
	c.Header("Content-Type", "text/html")
	c.String(200, output)
}
