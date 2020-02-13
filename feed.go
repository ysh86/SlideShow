package slideShow

import (
	"bytes"
	"fmt"
	"net/http"

	"github.com/mmcdole/gofeed"
	"github.com/yhat/scrape"
	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"
	"golang.org/x/net/html/charset"
)

type Feed struct {
	Pages  []*Page
	Images []*Image

	parser *gofeed.Parser
}

type Page struct {
	URL   string
	Title string
}

type Image struct {
	URL          string
	ThumbnailURL string

	Title          string
	Image          []byte
	ThumbnailImage []byte

	parent *Page
}

func NewFeed() (*Feed, error) {
	fp := gofeed.NewParser()
	if fp == nil {
		return nil, fmt.Errorf("cannot new")
	}
	return &Feed{parser: fp}, nil
}

func (f *Feed) ParseURLsAsync(urls []string) <-chan error {
	errc := make(chan error, 1)

	if f == nil {
		errc <- nil
		return errc
	}

	go func() {
		f.Pages = nil
		f.Images = nil
		for _, u := range urls {
			if u == "" {
				continue
			}

			feed, err := f.parser.ParseURL(u)
			if err != nil {
				//errc <- err
				continue
			}

			for _, item := range feed.Items {
				if item.Link == "" {
					continue
				}

				title := feed.Title
				if title != "" {
					if item.Title != "" {
						title += ": " + item.Title
					}
				} else {
					title = item.Title
				}

				page := &Page{
					URL:   item.Link,
					Title: title,
				}
				f.Pages = append(f.Pages, page)
			}
		}

		errc <- nil
	}()

	return errc
}

func (f *Feed) ParsePagesAsync(pages []*Page) <-chan error {
	errc := make(chan error, 1)

	if f == nil {
		errc <- nil
		return errc
	}

	go func() {
		f.Pages = nil
		f.Images = nil
		for _, page := range pages {
			if page.URL == "" {
				continue
			}

			// get
			resp, err := http.Get(page.URL)
			if err != nil {
				//errc <- err
				continue
			}
			defer resp.Body.Close()

			// charset
			r, err := charset.NewReader(resp.Body, resp.Header.Get("content-type"))
			if err != nil {
				//errc <- err
				continue
			}

			// parse
			root, err := html.Parse(r)
			if err != nil {
				//errc <- err
				continue
			}

			if page.Title == "" {
				title, ok := scrape.Find(root, func(n *html.Node) bool {
					return n.Type == html.ElementNode && n.DataAtom == atom.Title
				})
				if ok {
					page.Title = title.FirstChild.Data
				}
			}

			matcher := func(n *html.Node) bool {
				return n.DataAtom == atom.Img && n.Parent != nil && n.Parent.DataAtom == atom.A
			}
			imgs := scrape.FindAll(root, matcher)
			i := 0
			for _, img := range imgs {
				url := scrape.Attr(img.Parent, "href")
				thumburl := scrape.Attr(img, "src")
				if url == "" || thumburl == "" {
					//errc <- err
					continue
				}
				if !bytes.Contains([]byte(url), []byte("jpg")) &&
					!bytes.Contains([]byte(url), []byte("jpeg")) {
					continue
				}

				var title string
				if page.Title != "" {
					title = fmt.Sprintf("%s: %d", page.Title, i+1)
				} else {
					title = fmt.Sprintf("%d", i+1)
				}

				image := &Image{
					URL:          url,
					ThumbnailURL: thumburl,
					Title:        title,
					parent:       page,
				}

				f.Images = append(f.Images, image)
				i++
			}

			f.Pages = append(f.Pages, page)
		}

		errc <- nil
	}()

	return errc
}
