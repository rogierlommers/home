package greedy

import (
	"testing"
)

func TestScraping(t *testing.T) {

	tests := []struct {
		url           string
		expectedTitle string
	}{
		{"https://www.ah.nl/allerhande/recept/R-R1195866/tikka-masala-met-vegetarische-kipstukjes-en-bloemkoolrijst", "Tikka masala met vegetarische kipstukjes en bloemkoolrijst recept - Allerhande | Albert Heijn"},
		{"https://rogier.lommers.org", "Rogier Lommers - Résumé"},
		{"https://www.reddit.com/r/fuckaroundandfindout/comments/1o3o0z4/theres_a_speed_limit_for_a_reason/", "Reddit - The heart of the internet"},
	}

	for _, tt := range tests {

		// create instance of article that needs to be scraped
		g := &GreedyURL{URL: tt.url}

		// start scraping
		if err := g.scrape(); err != nil {
			t.Errorf("Scrape() error: %v", err)
			continue
		}

		// check if title matches expected title
		if g.Title != tt.expectedTitle {
			t.Errorf("For URL %q, expected title %q, got %q", tt.url, tt.expectedTitle, g.Title)
		}
	}

}

func TestGetBaseURL(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{
			input:    "https://example.com/path/to/page",
			expected: "example.com/path/to/page",
		},
		{
			input:    "http://sub.domain.com/",
			expected: "sub.domain.com/",
		},
		{
			input:    "ftp://ftp.example.com/files",
			expected: "ftp.example.com/files",
		},
		{
			input:    "not a url",
			expected: "not a url",
		},
		{
			input:    "",
			expected: "",
		},
	}

	for _, tt := range tests {
		got := getBaseURL(tt.input)
		if got != tt.expected {
			t.Errorf("getBaseURL(%q) = %q; want %q", tt.input, got, tt.expected)
		}
	}
}
