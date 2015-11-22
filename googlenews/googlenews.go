/*
Package googlenews handles downloading and parsing search results for Google News queries output as RSS.
*/
package googlenews

import (
	"errors"
	"fmt"
	"net/url"

	"github.com/sha1sum/groupme_news_bot/matchers"
)

// Link is used to house a string URL along with any error that may have resulted from fetching the link from
// Google News.
type Link struct {
	URL string
	Err error
}

// FirstLink takes a search term and queries Google News for results, then parses the first story's raw link from the
// RSS output returned by Google News.
func FirstLink(term string, c chan Link) {
	// Fetch the Google news search results for the search term as an RSS feed.
	doc, err := matchers.Retrieve("http://news.google.com/news?q=" + url.QueryEscape(term) + "&output=rss")
	if err != nil {
		c <- Link{Err: err}
		return
	}
	// Get the <item>'s from the feed.
	items := doc.Channel.Item
	// If there are no items, return a "No results" error.
	if len(items) < 1 {
		c <- Link{Err: errors.New("No results for \"" + term + "\".")}
		return
	}
	fmt.Println("Link retrieved.")
	// Get the link with all the Googley stuff in it
	c <- parseLink(doc.Channel.Item[0])
}

// parseLink takes an RSS <item> struct and parses the link to the original story from the Google link to the item.
func parseLink(item matchers.Item) Link {
	l := item.Link
	parsed, err := url.Parse(l)
	if err != nil {
		return Link{Err: err}
	}
	// Get the query string values so we can just get the normal URL instead of the Googley one.
	queryVals, err := url.ParseQuery(parsed.RawQuery)
	if err != nil {
		return Link{Err: err}
	}
	ls := queryVals["url"]
	if len(ls) < 1 {
		return Link{URL: item.Link}
	}
	fmt.Println("Found link", ls[0])
	return Link{URL: ls[0]}
}
