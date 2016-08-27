// Copyright 2016 Google Inc. All rights reserved.
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to writing, software distributed
// under the License is distributed on a "AS IS" BASIS, WITHOUT WARRANTIES OR
// CONDITIONS OF ANY KIND, either express or implied.
//
// See the License for the specific language governing permissions and
// limitations under the License.

package podcast2youtube

import (
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"io/ioutil"
	"strconv"

	"github.com/golang/freetype/truetype"
	"golang.org/x/image/font"
	"golang.org/x/image/math/fixed"
)

// CreateImage creates a new image containing the given logo and text.
func CreateImage(logo image.Image, text, fgHex, bgHex string, width, height int) (image.Image, error) {
	bg, err := hexToColor(bgHex)
	if err != nil {
		return nil, fmt.Errorf("invalid background color %q: %v", bgHex, err)
	}

	// create a new image with the given background color
	m := image.NewRGBA(image.Rect(0, 0, width, height))
	draw.Draw(m, m.Bounds(), image.NewUniform(bg), image.Point{}, draw.Src)

	pos := image.Point{(m.Bounds().Max.X - logo.Bounds().Max.X) / 2, m.Bounds().Max.Y / 3}
	draw.Draw(m, m.Bounds().Add(pos), logo, image.Point{}, draw.Over)

	// load a font
	f, err := loadFont("Roboto-Light.ttf")
	if err != nil {
		return nil, fmt.Errorf("could not load font: %v", err)
	}

	fg, err := hexToColor(fgHex)
	if err != nil {
		return nil, fmt.Errorf("invalid foreground color %q: %v", fgHex, err)
	}

	face, dot := fitFontSize(f, m.Bounds().Max.X, text)
	// and use it to write some text
	d := &font.Drawer{
		Dst:  m,
		Src:  image.NewUniform(fg),
		Face: face,
		Dot:  dot,
	}
	d.DrawString(text)
	return m, nil
}

// hexToColor parses a hexadecimal color and returns it as an RGBA color.
func hexToColor(s string) (*color.RGBA, error) {
	if len(s) != 6 {
		return nil, fmt.Errorf("color should be 6 digits")
	}
	n, err := strconv.ParseInt(s, 16, 64)
	if err != nil {
		return nil, fmt.Errorf("not hexadecimal: %v", err)
	}
	b, n := uint8(n%256), n/256
	g, n := uint8(n%256), n/256
	r, n := uint8(n%256), n/256
	return &color.RGBA{r, g, b, 255}, nil
}

// loadFont loads a truetype font in the given path.
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

// fitFontSize finds the font size at which the given text fits 80% of the
// given width and the indentation required to center the text.
func fitFontSize(f *truetype.Font, width int, text string) (font.Face, fixed.Point26_6) {
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
