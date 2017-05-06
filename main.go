package main

import (
	"fmt"
	"net/http"
	"os"
	"strings"

	"golang.org/x/net/html"
)

const (
	lyricSearchBaseURL = "http://search.azlyrics.com/search.php?q="
	debugMode          = true
)

type searchResults struct {
	PastAlbums  bool
	InBold      bool
	Links       []link
	CurrentLink *link
}

type link struct {
	Addr   string
	Title  string
	Artist string
}

// Method for pretty printing link objects
func (l link) String() string {
	return fmt.Sprintf("Linke to: %s by %s - %s", l.Title, l.Artist, l.Addr)
}

// Check that all fields on the link are filled out
func (l *link) isComplete() bool {
	if l.Addr == "" || l.Title == "" || l.Artist == "" {
		return false
	}
	return true
}

func main() {
	// Grab the search term from the command line arguments
	var searchTerm string
	if len(os.Args) > 1 {
		searchTerm = strings.Join(os.Args[1:], " ")
	} else {
		// By default (if no params were provided), search for Madness
		searchTerm = "Madness"
	}

	// Grab the search page from online
	resp, err := http.Get(lyricSearchBaseURL + searchTerm)
	if err != nil {
		panic("Failed to scrape azlyrics website")
	}
	// Print out the contents of the webpage we scraped
	// bytes, _ := ioutil.ReadAll(resp.Body)
	// fmt.Println("HTML:\n\n", string(bytes))
	// resp.Body.Close()

	// Use the HTML library to tokenize the response
	z := html.NewTokenizer(resp.Body)

	// Parse the search page
	results := parseSearchPage(z)

	fmt.Println(results)

}

// Parse the html results of the search page
func parseSearchPage(tree *html.Tokenizer) *searchResults {
	var results searchResults
	// Loop over the tree
	for {
		// Grab the next element in the tree
		tt := tree.Next()

		// Switch on the few different element types
		switch {
		case tt == html.ErrorToken:
			// End of the document, we're done
			debugf("Finished processing search page")
			return &results
		case tt == html.StartTagToken: // HTML opening tags
			// Get the data about this node
			t := tree.Token()

			// Handle link nodes
			if t.Data == "a" {
				results.processLink(t)
			}

			// Looking for link data and encountered a bold area
			if results.CurrentLink != nil && t.Data == "b" {
				results.InBold = true
			}
		case tt == html.TextToken: // Plain text in the HTMl doc
			// Grab the contents of this node
			t := tree.Token()

			// Process the plain text
			results.processText(t)
		}

	}
}

// Process plain text from the document
func (sr *searchResults) processText(t html.Token) {
	// This text is within a bold tag
	if sr.InBold {
		// Unset the flag
		sr.InBold = false

		// Use this at the title if it is not set yet
		if sr.CurrentLink.Title == "" {
			sr.CurrentLink.Title = t.String()
			// Use this at the artist if it is not set yet
		} else if sr.CurrentLink.Artist == "" {
			sr.CurrentLink.Artist = t.String()
			// All the fields should be filled in by now, if they are not, panic
			if !sr.CurrentLink.isComplete() {
				panic("Found artist before completing the link, page layout must have changed")
			}
			sr.Links = append(sr.Links, *sr.CurrentLink)
			sr.CurrentLink = nil
		} else {
			panic("Found one to many text fields, they probably changed their site")
		}
	}
}

// Process a <a> element on page
func (sr *searchResults) processLink(t html.Token) {
	// Loop over each one of the link's attributes
	for _, attr := range t.Attr {
		// Don't worry about anything else until all album links have passed
		if !sr.PastAlbums {
			// The "More Album Results" link
			if attr.Key == "href" && strings.Contains(attr.Val, "w=albums") {
				sr.PastAlbums = true
				debugf("Finished processing albums")
			}
		} else {
			// Identify the actual search result links
			if attr.Key == "target" && attr.Val == "_blank" {
				sr.addNewLink(t.String())
			}
		}
	}
}

// Queue up the next result link to be processed
func (sr *searchResults) addNewLink(linkAddr string) {
	// Make sure the previous link has finshed processing
	if sr.CurrentLink != nil {
		panic("Found the next link before finishing the previous one. They probably changed their site")
	}

	var newLink link
	newLink.Addr = linkAddr

	sr.CurrentLink = &newLink
}

// Helper function for only printing things while in debug mode
func debugf(strs ...string) {
	if debugMode {
		fmt.Println(strings.Join(strs, ""))
	}
}
