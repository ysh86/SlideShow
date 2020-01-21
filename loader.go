package main

import (
	"image"
	"image/color"
	"time"

	"golang.org/x/exp/shiny/screen"
	"golang.org/x/mobile/event/paint"

	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"

	_ "golang.org/x/image/bmp"
	_ "golang.org/x/image/tiff"
	_ "golang.org/x/image/webp"
)

type asyncLoader struct {
	cur chan image.Image

	src []image.Image
}

func NewAsyncLoader() (*asyncLoader, error) {
	cur := make(chan image.Image, 1)
	return &asyncLoader{cur: cur}, nil
}

func (l *asyncLoader) SetList() {
	/*
		if len(os.Args) < 2 {
			log.Fatal("no image file specified")
		}
		src, err := decode(os.Args[1])
		if err != nil {
			log.Fatal(err)
		}
	*/
	l.src = []image.Image{
		Checker.Gen(1280, 720, 16*5),
		nil, // URL0
		Checker.Gen(720, 1280, 16*5),
		nil, // URL1
		Checker.Gen(720, 720, 16*5),
		nil, // URL2
		nil, // URL3
		nil, // URL4
		nil, // URL5
	}
}

func loadAsync(done <-chan interface{}, src []image.Image) <-chan (<-chan image.Image) {
	ch := make(chan (<-chan image.Image), 4) // load in 4-para

	go func() {
		defer close(ch)
		for i, p := range src {
			loadch := make(chan image.Image, 1)

			go func(idx int, pic image.Image) {
				defer close(loadch)
				if pic != nil {
					loadch <- pic
					return
				}
				// TODO: load from URL
				duration := 1000 * time.Hour
				select {
				case <-time.After(duration):
					// loaded
					loadch <- &image.Uniform{color.White}
				case <-done:
					//log.Println("canceled loading:", idx)
					return
				}
			}(i, p)

			select {
			case ch <- loadch:
			case <-done:
				//log.Println("canceled queueing:", i)
				return
			}
		}
	}()

	return ch
}

func (l *asyncLoader) Run(done <-chan interface{}, w screen.Window) {
	ch := loadAsync(done, l.src)

	go func() {
		// keep blank pic
		duration := 500 * time.Millisecond
		select {
		case <-time.After(duration):
		case <-done:
			return
		}

		for loadch := range ch {
			// load next pic
			timeout := 2 * time.Second
			var next image.Image
			var ok bool
			select {
			case next, ok = <-loadch:
				if !ok {
					// skip this pic
					continue
				}
			case <-time.After(timeout):
				// skip this pic
				//log.Println("skip")
				continue
			case <-done:
				return
			}

			// paint next pic
			timeout = 1 * time.Second
		retryPaint:
			select {
			case l.cur <- next:
			case <-time.After(timeout):
				// paint Q is full?
				goto retryPaint
			case <-done:
				return
			}

			// paint in the UI thread
			w.Send(paint.Event{})
			duration := 4 * time.Second
			select {
			case <-time.After(duration):
			case <-done:
				return
			}
		}

		// signal EOF?
		//close(l.cur)
		//w.Send(paint.Event{})
	}()
}
