package main

import (
	"fmt"
	"os"

	"github.com/ysh86/slideShow"
)

func main() {
	feed, _ := slideShow.NewFeed()

	urls := os.Args[1:]
	errc := feed.ParseURLsAsync(urls)
	select {
	case err := <-errc:
		if err != nil {
			panic(err)
		}
	}
	total := len(feed.Pages)
	for i, page := range feed.Pages {
		fmt.Printf("%d/%d: %v, %v\n", i+1, total, page.Title, page.URL)
	}
}
