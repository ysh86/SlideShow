package main

import (
	"fmt"
	"image"
	"image/draw"
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

func main() {
	log.SetFlags(log.Ldate | log.Ltime | log.Lmicroseconds | log.Lshortfile)
	log.Println("Start")

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
		Checker.Gen(1280, 720, 8),
	}

	idx := 0
	driver.Main(func(s screen.Screen) {
		w, err := s.NewWindow(&screen.NewWindowOptions{
			Title:  "Image viewer",
			Width:  src[idx].Bounds().Dx(), // / 2, // TODO: HiDPI
			Height: src[idx].Bounds().Dy(), // / 2, // TODO: HiDPI
		})
		if err != nil {
			log.Fatal(err)
		}
		defer w.Release()

		srcSize := image.Pt(src[idx].Bounds().Dx(), src[idx].Bounds().Dy())
		buf, err := s.NewBuffer(srcSize)
		if err != nil {
			log.Fatal(err)
		}
		defer buf.Release()

		draw.Draw(buf.RGBA(), buf.Bounds(), src[idx], src[idx].Bounds().Min, draw.Src)

		tex, err := s.NewTexture(srcSize)
		if err != nil {
			log.Fatal(err)
		}
		defer tex.Release()

		tex.Upload(image.Point{}, buf, buf.Bounds())

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
