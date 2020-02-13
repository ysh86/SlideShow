package slideShow

import (
	"image"
	"net/http"
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
	Cur chan image.Image

	src []string
}

func NewAsyncLoader() (*asyncLoader, error) {
	cur := make(chan image.Image, 1)
	return &asyncLoader{Cur: cur}, nil
}

func (l *asyncLoader) SetList(src []string) {
	l.src = src
}

func urlAsync(url string) <-chan image.Image {
	ch := make(chan image.Image, 1)

	go func() {
		defer close(ch)

		// debug
		if url == "0" {
			ch <- Checker.Gen(1280, 720, 16*5)
			return
		}
		if url == "1" {
			ch <- Checker.Gen(720, 1280, 16*5)
			return
		}

		resp, err := http.Get(url)
		if err != nil {
			return
		}
		defer resp.Body.Close()

		pic, _, err := image.Decode(resp.Body)
		if err != nil {
			return
		}

		ch <- pic
	}()

	return ch
}

func loadAsync(done <-chan interface{}, src []string) <-chan (<-chan image.Image) {
	ch := make(chan (<-chan image.Image), 4) // load in 4-para

	go func() {
		defer close(ch)
		for i, s := range src {
			loadch := make(chan image.Image, 1)

			go func(idx int, url string) {
				defer close(loadch)
				select {
				case pic, ok := <-urlAsync(url):
					if ok {
						// loaded
						loadch <- pic
					}
				case <-done:
					//log.Println("canceled loading:", idx)
					return
				}
			}(i, s)

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
			case l.Cur <- next:
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
