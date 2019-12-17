package main

import (
	"image"
	"log"
	"os"
)

type AsyncImage struct {
	image.Image
}

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
