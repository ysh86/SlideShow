package slideShow

import (
	"image"
	"image/color"

	"golang.org/x/exp/shiny/screen"
	"golang.org/x/image/draw"
)

type noEffect struct {
	canvas     screen.Buffer
	tex        screen.Texture
	backGround image.Image
}

func NewNoEffectRenderer() (*noEffect, error) {
	return &noEffect{}, nil
}

func (r *noEffect) Release() {
	r.tex.Release()
	r.canvas.Release()
}

func (r *noEffect) Init(s screen.Screen, size image.Point, backGroundColor color.Color) error {
	canvas, err := s.NewBuffer(size)
	if err != nil {
		return err
	}

	tex, err := s.NewTexture(size)
	if err != nil {
		canvas.Release()
		return err
	}

	r.canvas = canvas
	r.tex = tex
	r.backGround = &image.Uniform{backGroundColor}

	r.Render(nil /* no foreground */)
	return nil
}

func (r *noEffect) Render(src image.Image) {
	// background
	draw.Draw(r.canvas.RGBA(), r.canvas.Bounds(), r.backGround, image.Point{}, draw.Src)

	// foreground
	if src != nil && !src.Bounds().Empty() {
		// calc ratio
		fx := float64(r.canvas.Bounds().Dx()) / float64(src.Bounds().Dx())
		fy := float64(r.canvas.Bounds().Dy()) / float64(src.Bounds().Dy())
		dstMin := r.canvas.Bounds().Min
		dstMax := r.canvas.Bounds().Max
		if fx < fy {
			// resize
			dstMax.Y = dstMin.Y + int(float64(src.Bounds().Dy())*fx+0.5)
			// centering vertically
			diff := (r.canvas.Bounds().Dy() - (dstMax.Y - dstMin.Y)) / 2
			dstMin.Y += diff
			dstMax.Y += diff
		} else {
			// resize
			dstMax.X = dstMin.X + int(float64(src.Bounds().Dx())*fy+0.5)
			// centering horizontally
			diff := (r.canvas.Bounds().Dx() - (dstMax.X - dstMin.X)) / 2
			dstMin.X += diff
			dstMax.X += diff
		}

		// resize!
		//draw.NearestNeighbor.Scale(
		draw.CatmullRom.Scale(
			r.canvas.RGBA(),
			image.Rectangle{dstMin, dstMax},
			src,
			src.Bounds(),
			draw.Src, nil)
	}

	r.tex.Upload(image.Point{}, r.canvas, r.canvas.Bounds())
}

func (r *noEffect) Swap(w screen.Window) {
	w.Copy(image.Point{}, r.tex, r.tex.Bounds(), screen.Src, nil)
	w.Publish()
}
