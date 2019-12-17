package main

import (
	"image"
	"image/color"
	"math"

	"golang.org/x/exp/shiny/gesture"
	"golang.org/x/exp/shiny/iconvg"
	"golang.org/x/exp/shiny/materialdesign/icons"
	"golang.org/x/exp/shiny/widget/node"
	"golang.org/x/exp/shiny/widget/theme"
	"golang.org/x/image/draw"
	"golang.org/x/image/font"
	"golang.org/x/image/math/fixed"
)

const pad = 12
const iconSize = 32

func loadDirIcon() image.Image {
	bounds := image.Rect(0, 0, iconSize, iconSize)
	icon := image.NewRGBA(bounds)

	var raster iconvg.Rasterizer
	raster.SetDstImage(icon, bounds, draw.Over)
	iconvg.Decode(&raster, icons.FileFolder, nil)

	return icon
}

type button struct {
	node.LeafEmbed

	pad     int
	label   string
	onClick func()
}

func newButton(label string, onClick func()) *button {
	b := &button{pad: pad, label: label, onClick: onClick}
	b.Wrapper = b

	return b
}

func (b *button) Measure(t *theme.Theme, widthHint, heightHint int) {
	face := t.AcquireFontFace(theme.FontFaceOptions{})
	defer t.ReleaseFontFace(theme.FontFaceOptions{}, face)

	b.MeasuredSize.X = font.MeasureString(face, b.label).Ceil() + 2*b.pad
	b.MeasuredSize.Y = face.Metrics().Ascent.Ceil() + face.Metrics().Descent.Ceil() + 2*b.pad
}

func (b *button) PaintBase(ctx *node.PaintBaseContext, origin image.Point) error {
	b.Marks.UnmarkNeedsPaintBase()
	face := ctx.Theme.AcquireFontFace(theme.FontFaceOptions{})
	defer ctx.Theme.ReleaseFontFace(theme.FontFaceOptions{}, face)

	draw.Draw(ctx.Dst, b.Rect.Add(origin), theme.Foreground.Uniform(ctx.Theme), image.Point{}, draw.Src)
	d := font.Drawer{
		Dst:  ctx.Dst,
		Src:  theme.Background.Uniform(ctx.Theme),
		Face: face,
		Dot:  fixed.Point26_6{X: fixed.I(b.Rect.Min.X + b.pad), Y: fixed.I(b.Rect.Min.Y + face.Metrics().Ascent.Ceil() + b.pad)},
	}
	d.DrawString(b.label)

	return nil
}

func (b *button) OnInputEvent(e interface{}, origin image.Point) node.EventHandled {
	if ev, ok := e.(gesture.Event); ok {
		if ev.Type == gesture.TypeTap && b.onClick != nil {
			b.onClick()
		}

		return node.Handled
	}

	return node.NotHandled
}

type cell struct {
	node.LeafEmbed

	pad      int
	icon     *scaledImage
	iconSize int
	space    int
	label    string
	onClick  func()
}

func newCell(icon image.Image, space int, label string, onClick func()) *cell {
	scaledIcon := newScaledImage(icon)
	c := &cell{pad: pad, icon: scaledIcon, iconSize: iconSize, space: space, label: label, onClick: onClick}
	c.Wrapper = c

	return c
}

func (c *cell) Measure(t *theme.Theme, widthHint, heightHint int) {
	face := t.AcquireFontFace(theme.FontFaceOptions{})
	defer t.ReleaseFontFace(theme.FontFaceOptions{}, face)

	c.MeasuredSize.X = c.iconSize + c.space + font.MeasureString(face, c.label).Ceil() + 2*c.pad
	c.MeasuredSize.Y = face.Metrics().Ascent.Ceil() + face.Metrics().Descent.Ceil() + 2*c.pad
}

func (c *cell) PaintBase(ctx *node.PaintBaseContext, origin image.Point) error {
	c.Marks.UnmarkNeedsPaintBase()
	face := ctx.Theme.AcquireFontFace(theme.FontFaceOptions{})
	defer ctx.Theme.ReleaseFontFace(theme.FontFaceOptions{}, face)

	img := c.icon.Src
	if img != nil {
		ratio := float32(img.Bounds().Max.Y) / float32(img.Bounds().Max.X)
		if img.Bounds().Max.Y > img.Bounds().Max.X {
			ratio = float32(img.Bounds().Max.X) / float32(img.Bounds().Max.Y)
		}
		scaled := scaleImage(img, c.iconSize, int(float32(c.iconSize)*ratio))
		draw.Draw(ctx.Dst, c.Rect.Add(origin), scaled, image.Point{}, draw.Over)
	}

	d := font.Drawer{
		Dst:  ctx.Dst,
		Src:  theme.Foreground.Uniform(ctx.Theme),
		Face: face,
		Dot: fixed.Point26_6{X: fixed.I(c.Rect.Min.X + origin.X + c.iconSize + c.space),
			Y: fixed.I(c.Rect.Min.Y + origin.Y + face.Metrics().Ascent.Ceil())},
	}
	d.DrawString(c.label)

	return nil
}

func (c *cell) OnInputEvent(e interface{}, origin image.Point) node.EventHandled {
	if ev, ok := e.(gesture.Event); ok {
		if ev.Type == gesture.TypeTap && c.onClick != nil {
			c.onClick()
		}

		return node.Handled
	}

	return node.NotHandled
}

type scaledImage struct {
	node.LeafEmbed
	Src image.Image
}

func newScaledImage(src image.Image) *scaledImage {
	w := &scaledImage{
		Src: src,
	}
	w.Wrapper = w
	return w
}

func (w *scaledImage) Measure(t *theme.Theme, widthHint, heightHint int) {
	w.MeasuredSize.X = 640
	w.MeasuredSize.Y = 480
}

func (w *scaledImage) PaintBase(ctx *node.PaintBaseContext, origin image.Point) error {
	w.Marks.UnmarkNeedsPaintBase()

	wRect := w.Rect.Add(origin)
	width := wRect.Max.X - wRect.Min.X
	height := wRect.Max.Y - wRect.Min.Y

	checkers.resize(width, height)
	draw.Draw(ctx.Dst, wRect, checkers, checkers.Bounds().Min, draw.Src)
	if w.Src == nil {
		return nil
	}

	ratio := float32(w.Src.Bounds().Max.X) / float32(w.Src.Bounds().Max.Y)

	imgWidth := int(math.Min(float64(width), float64(w.Src.Bounds().Max.X)))
	imgHeight := int(float32(imgWidth) / ratio)

	if imgHeight > height {
		imgHeight = int(math.Min(float64(height), float64(w.Src.Bounds().Max.Y)))
		imgWidth = int(float32(imgHeight) * ratio)
	}

	scaled := scaleImage(w.Src, imgWidth, imgHeight)
	offset := image.Point{(imgWidth - width) / 2, (imgHeight - height) / 2}

	draw.Draw(ctx.Dst, wRect, scaled, offset, draw.Over)
	return nil
}

func (w *scaledImage) SetImage(img image.Image) {
	w.Src = img
	w.Mark(node.MarkNeedsPaintBase)

	refresh(w)
}

func scaleImage(src image.Image, width, height int) image.Image {
	ret := image.NewRGBA(image.Rect(0, 0, width, height))

	draw.ApproxBiLinear.Scale(ret, ret.Bounds(), src, src.Bounds(), draw.Src, nil)

	return ret
}

var checkers = &checkerImage{}

type checkerImage struct {
	bounds image.Rectangle
}

func (c *checkerImage) resize(width, height int) {
	c.bounds = image.Rectangle{image.Pt(0, 0), image.Pt(width, height)}
}

func (c *checkerImage) ColorModel() color.Model {
	return color.RGBAModel
}

func (c *checkerImage) Bounds() image.Rectangle {
	return c.bounds
}

func (c *checkerImage) At(x, y int) color.Color {
	xr := x / 10
	yr := y / 10

	if xr%2 == yr%2 {
		return color.RGBA{0xc0, 0xc0, 0xc0, 0xff}
	}
	return color.RGBA{0x99, 0x99, 0x99, 0xff}
}

func refresh(_ node.Node) {
	// Ideally we should refresh but this requires a reference to the window
	// win.Send(paint.Event{})
}
