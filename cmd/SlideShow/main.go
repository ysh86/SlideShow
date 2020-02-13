package main

import (
	"bufio"
	"fmt"
	"image"
	"image/color"
	"log"
	"os"
	"strings"

	"golang.org/x/exp/shiny/driver"
	"golang.org/x/exp/shiny/screen"
	"golang.org/x/exp/shiny/unit"
	"golang.org/x/mobile/event/key"
	"golang.org/x/mobile/event/lifecycle"
	"golang.org/x/mobile/event/paint"
	"golang.org/x/mobile/event/size"

	"github.com/ysh86/slideShow"
)

// Default window size
const (
	winWidth  = 1920
	winHeight = 1280
)

// UI colors
var backGroundColor = color.Gray{32}

func main() {
	log.SetFlags(log.Ldate | log.Ltime | log.Lmicroseconds | log.Lshortfile)

	// parse pages
	var pages []*slideShow.Page
	if len(os.Args) == 1 {
		// pipe
		scanner := bufio.NewScanner(os.Stdin)
		for scanner.Scan() {
			elms := strings.Split(scanner.Text(), ",")
			//fmt.Printf("%#v\n", elms)
			if len(elms) >= 2 {
				u := strings.Trim(elms[len(elms)-1], " ")
				t := strings.Trim(elms[len(elms)-2], " ")
				pages = append(pages, &slideShow.Page{URL: u, Title: t})
			}
		}
	} else {
		// args
		urls := os.Args[1:]
		for _, u := range urls {
			pages = append(pages, &slideShow.Page{URL: u})
		}
	}
	feed, _ := slideShow.NewFeed()
	errc := feed.ParsePagesAsync(pages)
	select {
	case err := <-errc:
		if err != nil {
			panic(err)
		}
	}
	// image URLs
	var src []string
	for _, image := range feed.Images {
		//fmt.Printf("  i%d: %v, %v, %v\n", i, image.Title, image.ThumbnailURL, image.URL)
		src = append(src, image.URL)
	}
	/*
		src = []string{
			"https://upload.wikimedia.org/wikipedia/commons/2/2c/Rotating_earth_%28large%29.gif",
			"0",
			"https://upload.wikimedia.org/wikipedia/commons/6/6b/Phalaenopsis_JPEG.jpg",
			"https://upload.wikimedia.org/wikipedia/commons/4/47/PNG_transparency_demonstration_1.png",
			"https://upload.wikimedia.org/wikipedia/commons/b/b2/Vulphere_WebP_OTAGROOVE_demonstration_2.webp",
			"1",
		}
	*/
	if len(src) == 0 {
		log.Println("No images")
		return
	}

	// src images
	loader, err := slideShow.NewAsyncLoader()
	if err != nil {
		log.Fatal(err)
	}
	loader.SetList(src)
	log.Println("Start loader")

	// renderer
	renderer, err := slideShow.NewNoEffectRenderer()
	if err != nil {
		log.Fatal(err)
	}
	log.Println("Start renderer")

	// UI loop
	driver.Main(func(s screen.Screen) {
		w, err := s.NewWindow(&screen.NewWindowOptions{
			Title:  "Image viewer",
			Width:  winWidth,  // / 2, // TODO: HiDPI
			Height: winHeight, // / 2, // TODO: HiDPI
		})
		if err != nil {
			log.Fatal(err)
		}
		defer w.Release()

		// start loader
		done := make(chan interface{})
		loader.Run(done, w)
		defer func() { close(done); log.Println("Done loader") }()

		// init renderer
		err = renderer.Init(s, image.Pt(winWidth, winHeight), backGroundColor)
		if err != nil {
			log.Fatal(err)
		}
		defer func() { renderer.Release(); log.Println("Done renderer") }()

		var sz size.Event
		for {
			e := w.NextEvent()

			// This print message is to help programmers learn what events this
			// example program generates. A real program shouldn't print such
			// messages; they're not important to end users.
			format := "got %#v\n"
			if _, ok := e.(fmt.Stringer); ok {
				format = "got %v\n"
			}
			log.Printf(format, e)

			switch e := e.(type) {
			case lifecycle.Event:
				if e.To == lifecycle.StageDead {
					return
				}

			case key.Event:
				if e.Code == key.CodeEscape {
					return
				}

			case paint.Event:
				select {
				case cur, ok := <-loader.Cur:
					if ok {
						renderer.Render(cur)
					} else {
						// EOF
						renderer.Render(nil)
						log.Println("paint EOF")
					}
				default:
					// nothing to do
				}
				renderer.Swap(w)

			case size.Event:
				sz = e
				log.Printf("size: %#v, PointsPerInch: %#v\n", sz, unit.PointsPerInch)
				// TODO: resize canvas in renderer
				w.Send(paint.Event{})

			case error:
				log.Print(e)
			}
		}
	})

}
