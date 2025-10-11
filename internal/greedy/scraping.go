package greedy

import (
	"time"

	"github.com/sirupsen/logrus"
)

func (g *Greedy) scheduleScraping() {

	// Schedule scraping every scrapingFrequncy seconds
	go func() {
		for {
			logrus.Debugf("Starting scheduled scraping of articles")
			time.Sleep(time.Duration(scrapingFrequncy) * time.Second)

			onlyScrapedURLs := false
			urls, err := g.getURLs(onlyScrapedURLs)
			if err != nil {
				logrus.Errorf("Error fetching articles for scraping: %v", err)
				continue
			}

			for _, url := range urls {
				logrus.Infof("Scraping URL: %s", url.URL)
				if err := url.scrape(); err != nil {
					logrus.Errorf("Error scraping URL %s: %v", url.URL, err)
					continue
				}

				if err := g.updateURL(url); err != nil {
					logrus.Errorf("Error updating URL %s in database: %v", url.URL, err)
					continue
				}

				logrus.Debugf("Successfully scraped and updated URL: %s with title: %s", url.URL, url.Title)
			}

			logrus.Debugf("Completed scheduled scraping of articles")
		}
	}()

}
