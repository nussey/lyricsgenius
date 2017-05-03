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
)

func main() {
	var searchTerm string
	if len(os.Args) > 1 {
		searchTerm = strings.Join(os.Args[1:], " ")
	} else {
		searchTerm = "Madness"
	}

	resp, _ := http.Get(lyricSearchBaseURL + searchTerm)
	// Print out the contents of the webpage we scraped
	// bytes, _ := ioutil.ReadAll(resp.Body)
	// fmt.Println("HTML:\n\n", string(bytes))
	// resp.Body.Close()

	z := html.NewTokenizer(resp.Body)

	pastAlbums := false
	foo := false
	bar := false
	count := 0

	for {
		tt := z.Next()

		switch {
		case tt == html.ErrorToken:
			// End of the document, we're done
			return
		case tt == html.StartTagToken:
			t := z.Token()

			if t.Data == "a" {
				processLink(t)
			}

			if foo && t.Data == "b" {
				count++
				bar = true
			}
		case tt == html.TextToken:
			if bar {
				fmt.Println(z.Token())
				bar = false
				if count >= 2 {
					count = 0
					foo = false
				}
			}
		}
	}
}

func processLink(t html.Token) {
	for _, attr := range t.Attr {
		if !pastAlbums {
			if attr.Key == "href" && strings.Contains(attr.Val, "search") {
				pastAlbums = true
				fmt.Println("Parsed albums!")
			}
		} else {
			if attr.Key == "target" && attr.Val == "_blank" {
				// Keep track of which link we hit last, exit this loop once we find the first one in this block
				fmt.Println(t.String())
				fmt.Println(t.Attr)
				fmt.Println("We found a link!")
				fmt.Println("-----")

				foo = true
			}
		}
	}
}
