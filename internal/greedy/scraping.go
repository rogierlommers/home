package greedy

import (
	"time"

	"github.com/sirupsen/logrus"
)

func (g *Greedy) scheduleScraping() {

	// Schedule scraping every scrapingFrequncy seconds
	go func() {
		logrus.Infof("starting scraping every %d seconds", g.cfg.GreedyScrapingFrequency)

		for {
			logrus.Debugf("starting scheduled scraping of articles")
			time.Sleep(time.Duration(g.cfg.GreedyScrapingFrequency) * time.Second)

			onlyScrapedURLs := false
			urls, err := g.getURLs(onlyScrapedURLs)
			if err != nil {
				logrus.Errorf("error fetching articles for scraping: %v", err)
				continue
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

			logrus.Debugf("completed scheduled scraping of articles")
		}
	}()

}
