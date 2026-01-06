package greedy

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// performScraping executes the scraping logic for all unscraped URLs
func (g *Greedy) performScraping() error {
	onlyScrapedURLs := false
	urls, err := g.getURLs(onlyScrapedURLs)
	if err != nil {
		logrus.Errorf("error fetching articles for scraping: %v", err)
		return err
	}

	for _, url := range urls {
		logrus.Debugf("Scraping URL: %s", url.URL)

		if err := url.scrape(); err != nil {
			url.ScrapeCount++
			if err := g.updateURL(url, false); err != nil {
				logrus.Errorf("error updating URL %s in database: %v", url.URL, err)
			}
			continue
		}

		if err := g.updateURL(url, true); err != nil {
			logrus.Errorf("error updating URL %s in database: %v", url.URL, err)
			continue
		}

		logrus.Debugf("successfully scraped and updated URL: %s with title: %s", url.URL, url.Title)
	}

	return nil
}

func (g *Greedy) scheduleScraping() {

	// Schedule scraping every scrapingFrequncy seconds
	go func() {
		logrus.Infof("starting scraping every %d seconds", g.cfg.GreedyScrapingFrequency)

		for {
			logrus.Debugf("starting scheduled scraping of articles")
			time.Sleep(time.Duration(g.cfg.GreedyScrapingFrequency) * time.Second)

			if err := g.performScraping(); err != nil {
				logrus.Errorf("error during scheduled scraping: %v", err)
			}

			logrus.Debugf("completed scheduled scraping of articles")
		}
	}()

}

func (g Greedy) triggerScraping(c *gin.Context) {

	// trigger scraping
	go g.performScraping()

	// build html
	output := `<!DOCTYPE html>
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
		  <p><em>Succesfully triggered scraping.</em></p>
		</section>

	  </main>
	
	</body>
	
	</html>`

	// serve
	c.Header("Content-Type", "text/html")
	c.String(200, output)
}
