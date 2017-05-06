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

func (l *link) isComplete() bool {
	if l.Addr == "" || l.Title == "" || l.Artist == "" {
		return false
	}
	return true
}

func main() {
	var searchTerm string
	if len(os.Args) > 1 {
		searchTerm = strings.Join(os.Args[1:], " ")
	} else {
		searchTerm = "Madness"
	}

	resp, err := http.Get(lyricSearchBaseURL + searchTerm)
	if err != nil {
		panic("Failed to scrape azlyrics website")
	}
	// Print out the contents of the webpage we scraped
	// bytes, _ := ioutil.ReadAll(resp.Body)
	// fmt.Println("HTML:\n\n", string(bytes))
	// resp.Body.Close()

	z := html.NewTokenizer(resp.Body)

	results := parseSearchPage(z)

	fmt.Println(results)

}

func parseSearchPage(tree *html.Tokenizer) *searchResults {
	var results searchResults
	for {
		tt := tree.Next()

		switch {
		case tt == html.ErrorToken:
			// End of the document, we're done
			return &results
		case tt == html.StartTagToken:
			t := tree.Token()

			if t.Data == "a" {
				results.processLink(t)
			}

			if results.CurrentLink != nil && t.Data == "b" {
				fmt.Println("foobar2")
				results.InBold = true
			}
		case tt == html.TextToken:
			t := tree.Token()

			results.processText(t)
		}

	}
}

func (sr *searchResults) processText(t html.Token) {
	if sr.InBold {
		sr.InBold = false
		fmt.Println(t.String())

		if sr.CurrentLink.Title == "" {
			sr.CurrentLink.Title = t.String()
		} else if sr.CurrentLink.Artist == "" {
			sr.CurrentLink.Artist = t.String()
			fmt.Println(sr.CurrentLink)
			sr.Links = append(sr.Links, *sr.CurrentLink)
			sr.CurrentLink = nil
		} else {
			panic("We found one to many text fields, they probably changed their site")
		}
	}
}

// Process a <a> element on page
func (sr *searchResults) processLink(t html.Token) {
	// Loop over each one of the link's attributes
	for _, attr := range t.Attr {
		// If we haven't made it past albums yet, we won't find any real links
		if !sr.PastAlbums {
			// The "More Album Results" link
			if attr.Key == "href" && strings.Contains(attr.Val, "w=albums") {
				sr.PastAlbums = true
				debugf("Finished processing past albums")
			}
		} else {
			// If we have found one of the links we care about
			if attr.Key == "target" && attr.Val == "_blank" {
				sr.addNewLink(t.String())
			}
		}
	}
}

func (sr *searchResults) addNewLink(linkAddr string) {
	if sr.CurrentLink != nil {
		panic("We found the next link before finishing the previous one. They probably changed their site")
	}

	var newLink link
	newLink.Addr = linkAddr

	sr.CurrentLink = &newLink
}

func debugf(strs ...string) {
	if debugMode {
		fmt.Println(strings.Join(strs, ""))
	}
}
