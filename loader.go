package main

import (
	"image"
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
	idx int
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
		Checker.Gen(720, 1280, 16*5),
		Checker.Gen(720, 720, 16*5),
	}
	l.idx = 0
}

func (l *asyncLoader) Run(w screen.Window, done <-chan struct{}) {
	len := len(l.src)
	if len <= 0 {
		return
	}

	go func() {
		duration := 500 * time.Millisecond
		for {
			select {
			case <-time.After(duration):
			case <-done:
				return
			}
			select {
			case l.cur <- l.src[l.idx]:
				l.idx = (l.idx + 1) % len
				duration = 4 * time.Second
			default:
				// Q is full? Wait one more second.
				duration = 1 * time.Second
			}
			w.Send(paint.Event{})
		}
	}()
}
