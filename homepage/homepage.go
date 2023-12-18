package homepage

import (
	"github.com/gin-gonic/gin"
	"github.com/rogierlommers/home/config"
)

func Add(router *gin.Engine, cfg config.AppConfig) {
	router.GET("/", displayHome)
}

func displayHome(c *gin.Context) {

	output := `<!DOCTYPE html>
	<html lang="en">
	
	<head>
	  <meta charset="utf-8" />
	  <meta name="viewport" content="width=device-width, initial-scale=1" />
	  <title>home service</title>
	  <link rel="stylesheet" href="https://fonts.googleapis.com/css?family=Roboto:300,300italic,700,700italic" />
	  <link rel="stylesheet" href="https://cdnjs.cloudflare.com/ajax/libs/normalize/8.0.1/normalize.css" />
	  <link rel="stylesheet" href="https://cdnjs.cloudflare.com/ajax/libs/milligram/1.4.1/milligram.min.css" />
	  <link rel="stylesheet" href="https://milligram.io/styles/main.css" />
	</head>
	
	<body>
	  <main class="wrapper">
	
		<section class="container" id="examples">
		  <h1 class="title">home service</h1>
		  <p>
		    <em>Everything you need at home...</em>
			<ul>
				<li><a href="/metrics">prometheus metrics</a></li>
				<li><a href="/api/greedy/rss">greedy rss url</a></li>
			</ul>
		  </p>
		</section>

	  </main>
	
	</body>
	
	</html>`

	// serve
	c.Header("Content-Type", "text/html")
	c.String(200, output)
}
