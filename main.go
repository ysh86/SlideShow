package main

import (
	"fmt"
	"image"
	"log"
	"os"

	"golang.org/x/exp/shiny/driver"
	"golang.org/x/exp/shiny/screen"
	"golang.org/x/exp/shiny/widget"

	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"

	_ "golang.org/x/image/bmp"
	_ "golang.org/x/image/tiff"
	_ "golang.org/x/image/webp"
)

func decode(filename string) (image.Image, error) {
	f, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	m, _, err := image.Decode(f)
	if err != nil {
		return nil, fmt.Errorf("could not decode %s: %v", filename, err)
	}
	return m, nil
}

func main() {
	log.SetFlags(log.Ldate | log.Ltime | log.Lmicroseconds | log.Lshortfile)
	log.Println("Start")

	driver.Main(func(s screen.Screen) {
		if len(os.Args) < 2 {
			log.Fatal("no image file specified")
		}
		src, err := decode(os.Args[1])
		if err != nil {
			log.Fatal(err)
		}
		w := widget.NewSheet(widget.NewImage(src, src.Bounds()))
		log.Printf("Window: %dx%d\n", src.Bounds().Dx(), src.Bounds().Dy())
		if err := widget.RunWindow(s, w, &widget.RunWindowOptions{
			NewWindowOptions: screen.NewWindowOptions{
				Width:  src.Bounds().Dx(),
				Height: src.Bounds().Dy(),
				Title:  "ImageView Shiny Example",
			},
		}); err != nil {
			log.Fatal(err)
		}
	})

}
