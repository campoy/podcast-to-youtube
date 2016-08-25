package main

import (
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"io/ioutil"
	"os"

	"github.com/golang/freetype/truetype"
	"golang.org/x/image/font"
	"golang.org/x/image/math/fixed"
)

func createImage(number, title string) (image.Image, error) {
	text := fmt.Sprintf("episode #%v: %s", number, title)

	// create a new image with the given background color
	m := image.NewRGBA(image.Rect(0, 0, 1200, 800))
	bg := image.NewUniform(color.RGBA{0, 150, 136, 255})
	draw.Draw(m, m.Bounds(), bg, image.Point{}, draw.Src)

	// add the logo
	logo, err := loadImage("logo.png")
	if err != nil {
		return nil, fmt.Errorf("could not load logo: %v", err)
	}
	pos := image.Point{(m.Bounds().Max.X - logo.Bounds().Max.X) / 2, m.Bounds().Max.Y / 3}
	draw.Draw(m, m.Bounds().Add(pos), logo, image.Point{}, draw.Over)

	// load a font
	f, err := loadFont("Roboto-Light.ttf")
	if err != nil {
		return nil, fmt.Errorf("could not load font: %v", err)
	}

	face, dot := fitFace(f, m.Bounds().Max.X, text)
	// and use it to write some text
	d := &font.Drawer{
		Dst:  m,
		Src:  image.White,
		Face: face,
		Dot:  dot,
	}
	d.DrawString(text)
	return m, nil
}

func loadImage(path string) (image.Image, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("could not open %s: %v", path, err)
	}
	defer f.Close()
	m, err := png.Decode(f)
	if err != nil {
		return nil, fmt.Errorf("could not decode %s: %v", path, err)
	}
	return m, nil
}

func loadFont(path string) (*truetype.Font, error) {
	b, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("could not open file: %v", err)
	}

	f, err := truetype.Parse(b)
	if err != nil {
		return nil, fmt.Errorf("could not parse font: %v", err)
	}
	return f, nil
}

func fitFace(f *truetype.Font, width int, text string) (font.Face, fixed.Point26_6) {
	// fit 80% of the width
	fixw := fixed.I(int(0.8 * float64(width)))

	curWidth := 2 * fixw
	var face font.Face

	// find which is the good font size for the width
	for size := 100.0; curWidth > fixw; size-- {
		face = truetype.NewFace(f, &truetype.Options{
			Size:    size,
			Hinting: font.HintingNone,
			DPI:     72,
		})
		curWidth = (&font.Drawer{Face: face}).MeasureString(text)
	}

	// and center the result
	dot := fixed.Point26_6{X: (fixed.I(width) - curWidth) / 2, Y: fixed.I(500)}
	return face, dot
}
