package main

import (
	"flag"
	"fmt"
	"image"
	"io/ioutil"
	"log"
	"os"
	"path"
	"path/filepath"
	"strings"

	"golang.org/x/exp/shiny/driver"
	"golang.org/x/exp/shiny/screen"
	"golang.org/x/exp/shiny/unit"
	"golang.org/x/exp/shiny/widget"
	"golang.org/x/exp/shiny/widget/node"
	"golang.org/x/exp/shiny/widget/theme"
	"golang.org/x/mobile/event/lifecycle"

	"image/color"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"

	_ "golang.org/x/image/bmp"
	_ "golang.org/x/image/tiff"
	_ "golang.org/x/image/webp"
)

const space = 10

var padSize = unit.DIPs(space * 2)
var spaceSize = unit.DIPs(space)

var index = 0
var gmainView *scaledImage
var gname *widget.Label
var images []*AsyncImage
var names []string

func changeImage(offset int) {
	newidx := index + offset
	if newidx < 0 || newidx >= len(images) {
		return
	}

	chooseImage(newidx)
}

func previousImage() {
	changeImage(-1)
}

func nextImage() {
	changeImage(1)
}

func chooseImage(idx int) {
	index = idx
	gmainView.SetImage(images[idx])

	gname.Text = names[idx]
	gname.Mark(node.MarkNeedsMeasureLayout)
	gname.Mark(node.MarkNeedsPaintBase)
	refresh(gname)
}

func expandSpace() node.Node {
	return widget.WithLayoutData(widget.NewSpace(),
		widget.FlowLayoutData{ExpandAlong: true, ExpandAcross: true, AlongWeight: 1})
}

func makeBar() node.Node {
	prev := newButton("Previous", previousImage)
	next := newButton("Next", nextImage)
	gname = widget.NewLabel("Filename")

	flow := widget.NewFlow(widget.AxisHorizontal,
		prev,
		expandSpace(),
		widget.NewPadder(widget.AxisBoth, padSize, gname),
		expandSpace(),
		next)

	bar := widget.NewUniform(theme.Neutral, flow)

	return widget.WithLayoutData(bar,
		widget.FlowLayoutData{ExpandAlong: true, ExpandAcross: true})
}

func makeCell(idx int, name string) *cell {
	var onClick func()
	var icon image.Image
	if idx < 0 {
		icon = loadDirIcon()
	} else {
		onClick = func() { chooseImage(idx) }
	}

	return newCell(icon, space, name, onClick)
}

func makeList(dir string) node.Node {
	var fileNames []string
	fileInfos, _ := ioutil.ReadDir(dir)
	for _, info := range fileInfos {
		if info.IsDir() {
			continue
		}

		ext := strings.ToLower(filepath.Ext(info.Name()))
		if ext == ".jpg" || ext == ".jpeg" || ext == ".png" || ext == ".gif" ||
			ext == ".bmp" || ext == ".tiff" || ext == ".webp" {
			fileNames = append(fileNames, info.Name())
		}
	}

	parent := makeCell(-1, filepath.Base(dir))
	children := []node.Node{parent}

	for idx, name := range fileNames {
		cell := makeCell(idx, name)
		i := idx
		aimg := NewAsyncImage(path.Join(dir, name), func(path string, img image.Image) {
			log.Printf("loaded: %d: %s\n", i, path)
			cell.icon.SetImage(img)
			if i == index {
				gmainView.SetImage(img)
			}
		})
		children = append(children, cell)

		images = append(images, aimg)
		names = append(names, name)
	}

	return widget.NewFlow(widget.AxisVertical, children...)
}

type myRoot struct {
	*widget.Sheet
}

func (r *myRoot) Measure(t *theme.Theme, widthHint, heightHint int) {
	log.Printf("root: Measure: DPI=%f, WxH=%dx%d\n", t.DPI, widthHint, heightHint)
	r.Sheet.Measure(t, widthHint, heightHint)
}
func (r *myRoot) Paint(ctx *node.PaintContext, origin image.Point) error {
	log.Printf("root: Paint: DPI=%f, screen=%v, drawer=%v, origin=%s\n", ctx.Theme.DPI, ctx.Screen, ctx.Drawer, origin)
	return r.Sheet.Paint(ctx, origin)
}
func (r *myRoot) OnLifecycleEvent(e lifecycle.Event) {
	log.Printf("root: OnLifecycleEvent: %s\n", e)
	r.Sheet.OnLifecycleEvent(e)
}

func (r *myRoot) OnInputEvent(e interface{}, origin image.Point) node.EventHandled {
	log.Printf("root: OnInputEvent: event=%s, origin=%s\n", e, origin)
	return r.Sheet.OnInputEvent(e, origin)
}

func loadUI(dir string) {
	driver.Main(func(s screen.Screen) {
		list := makeList(dir)

		var img image.Image
		gmainView = newScaledImage(img)
		scaledImage := widget.WithLayoutData(
			gmainView,
			widget.FlowLayoutData{ExpandAlong: true, ExpandAcross: true, AlongWeight: 4})

		body := widget.NewFlow(widget.AxisHorizontal,
			list,
			widget.NewPadder(widget.AxisHorizontal, spaceSize, nil),
			scaledImage)
		expanding := widget.WithLayoutData(
			widget.NewPadder(widget.AxisBoth, padSize, body),
			widget.FlowLayoutData{ExpandAlong: true, ExpandAcross: true, AlongWeight: 4})

		container := widget.NewFlow(widget.AxisVertical,
			makeBar(),
			expanding)
		sheet := widget.NewSheet(widget.NewUniform(theme.Dark, container))
		mySheet := &myRoot{sheet}

		if len(images) > 0 {
			chooseImage(0)
		}

		var th theme.Theme
		th.Palette = &theme.Palette{
			theme.Light:      image.Uniform{C: color.RGBA{0xf5, 0xf5, 0xf5, 0xff}}, // Material Design "Grey 100".
			theme.Neutral:    image.Uniform{C: color.RGBA{0xee, 0xee, 0xee, 0xff}}, // Material Design "Grey 200".
			theme.Dark:       image.Uniform{C: color.RGBA{0xe0, 0xe0, 0xe0, 0xff}}, // Material Design "Grey 300".
			theme.Accent:     image.Uniform{C: color.RGBA{0x21, 0x96, 0xf3, 0xff}}, // Material Design "Blue 500".
			theme.Foreground: image.Uniform{C: color.RGBA{0x00, 0x00, 0x00, 0xff}}, // Material Design "Black".
			theme.Background: image.Uniform{C: color.RGBA{0xff, 0xff, 0xff, 0xff}}, // Material Design "White".
		}
		container.Measure(&th, 0, 0)
		if err := widget.RunWindow(s, mySheet, &widget.RunWindowOptions{
			NewWindowOptions: screen.NewWindowOptions{
				Title:  "GoImages",
				Width:  container.MeasuredSize.X,
				Height: container.MeasuredSize.Y,
			},
			Theme: th,
		}); err != nil {
			log.Fatal(err)
		}
	})
}

func main() {
	log.SetFlags(log.Ldate | log.Ltime | log.Lmicroseconds | log.Lshortfile)

	flag.Usage = func() {
		fmt.Fprintln(os.Stderr, "goimages takes a single, optional, directory parameter")
	}
	flag.Parse()

	dir, _ := os.Getwd()
	if len(flag.Args()) > 1 {
		flag.Usage()
		os.Exit(2)
	} else if len(flag.Args()) == 1 {
		dir = flag.Args()[0]

		if _, err := ioutil.ReadDir(dir); os.IsNotExist(err) {
			fmt.Fprintln(os.Stderr, "Directory", dir, "does not exist or could not be read")
			os.Exit(1)
		}
	}

	log.Printf("start: %s\n", dir)
	loadUI(dir)
}
