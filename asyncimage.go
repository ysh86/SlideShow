package main

import (
	"image"
	"image/color"
	"log"
	"os"
)

// AsyncImage is a async image loader.
type AsyncImage struct {
	image.Image
}

// NewAsyncImage returns a new AsyncImage and starts loading.
func NewAsyncImage(path string, loaded func(string, image.Image)) *AsyncImage {
	aimg := &AsyncImage{}
	go aimg.load(path, loaded)

	return aimg
}

func (a *AsyncImage) load(path string, loaded func(string, image.Image)) {
	f, err := os.Open(path)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	a.Image, _, err = image.Decode(f)
	if err != nil {
		log.Fatal(err)
	}

	loaded(path, a)
}

// Checker is a checker image generator.
var Checker = &checkerImage{}

type checkerImage struct {
	bounds image.Rectangle
	block  int
}

func (c *checkerImage) Gen(width, height, block int) image.Image {
	return &checkerImage{
		bounds: image.Rectangle{image.Pt(0, 0), image.Pt(width, height)},
		block:  block,
	}
}

func (c *checkerImage) At(x, y int) color.Color {
	xr := x / c.block
	yr := y / c.block

	if xr&1 == yr&1 {
		return color.RGBA{0xc0, 0xc0, 0xc0, 0xff}
	}
	return color.RGBA{0x99, 0x99, 0x99, 0xff}
}

func (c *checkerImage) Bounds() image.Rectangle {
	return c.bounds
}

func (c *checkerImage) ColorModel() color.Model {
	return color.RGBAModel
}
