package main

import (
	"fmt"
	"image"
	"image/draw"
	"image/color"
	"log"

	"golang.org/x/exp/shiny/driver"
	"golang.org/x/exp/shiny/screen"
	"golang.org/x/exp/shiny/unit"
	"golang.org/x/mobile/event/key"
	"golang.org/x/mobile/event/lifecycle"
	"golang.org/x/mobile/event/paint"
	"golang.org/x/mobile/event/size"

	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"

	_ "golang.org/x/image/bmp"
	_ "golang.org/x/image/tiff"
	_ "golang.org/x/image/webp"
)

// Default window size
const (
	WinWidth = 1920
	WinHeight = 1280
)

func main() {
	log.SetFlags(log.Ldate | log.Ltime | log.Lmicroseconds | log.Lshortfile)
	log.Println("Start")

	// images
	/*
		if len(os.Args) < 2 {
			log.Fatal("no image file specified")
		}
		src, err := decode(os.Args[1])
		if err != nil {
			log.Fatal(err)
		}
	*/
	src := []image.Image{
		Checker.Gen(1280, 720, 16),
		Checker.Gen(720, 1280, 16),
	}
	idx := 1

	driver.Main(func(s screen.Screen) {
		w, err := s.NewWindow(&screen.NewWindowOptions{
			Title:  "Image viewer",
			Width:  WinWidth, // / 2, // TODO: HiDPI
			Height: WinHeight, // / 2, // TODO: HiDPI
		})
		if err != nil {
			log.Fatal(err)
		}
		defer w.Release()

		canvasSize := image.Pt(WinWidth, WinHeight)
		canvas, err := s.NewBuffer(canvasSize)
		if err != nil {
			log.Fatal(err)
		}
		defer canvas.Release()

		tex, err := s.NewTexture(canvasSize)
		if err != nil {
			log.Fatal(err)
		}
		defer tex.Release()

		draw.Draw(canvas.RGBA(), canvas.Bounds(), &image.Uniform{color.Gray{32}}, image.Point{}, draw.Src)
		// TODO: resize & centoring
		draw.Draw(canvas.RGBA(), canvas.Bounds(), src[idx], src[idx].Bounds().Min, draw.Src)
		tex.Upload(image.Point{}, canvas, canvas.Bounds())

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
				w.Copy(image.Point{}, tex, tex.Bounds(), screen.Src, nil)
				w.Publish()

			case size.Event:
				sz = e
				log.Printf("size: %#v, PointsPerInch: %#v\n", sz, unit.PointsPerInch)
				w.Send(paint.Event{})

			case error:
				log.Print(e)
			}
		}
	})

}
