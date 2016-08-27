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

// Package image generates images containing a logo and some text below.
package image

import (
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"io/ioutil"
	"os"

	// This registers the supported formats for image.Decode.
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"

	"github.com/golang/freetype/truetype"
	"golang.org/x/image/font"
	"golang.org/x/image/math/fixed"
)

// Params contains the parameters that describe an image.
// They are all required for image creation.
type Params struct {
	Logo       string      // Filepath to a logo for the top half of the image.
	Text       string      // Text to display below the logo.
	Font       string      // Filepath to the TrueType font used for the text.
	Foreground color.Color // Color for the text.
	Background color.Color // Color for the background.
	Width      int         // Width of the image in pixels.
	Height     int         // Height of the image in pixels.
}

// Generate generates a new image given the corresponding parameters.
func Generate(p Params) (image.Image, error) {
	// We create a new image with the given background color.
	m := image.NewRGBA(image.Rect(0, 0, p.Width, p.Height))
	draw.Draw(m, m.Bounds(), image.NewUniform(p.Background), image.Point{}, draw.Src)

	logo, err := loadImg(p.Logo)
	if err != nil {
		return nil, fmt.Errorf("could not open %s: %v", p.Logo, err)
	}
	pos := image.Point{(m.Bounds().Max.X - logo.Bounds().Max.X) / 2, m.Bounds().Max.Y / 3}
	draw.Draw(m, m.Bounds().Add(pos), logo, image.Point{}, draw.Over)

	// Then load the font to be used with the text.
	f, err := loadFont(p.Font)
	if err != nil {
		return nil, fmt.Errorf("could not load font: %v", err)
	}

	// We leave a padding around the text by fitting the width to only 80% of the image width.
	paddedWidth := int(0.8 * float64(m.Bounds().Max.X))
	face, textWidth := fitFontSize(f, paddedWidth, p.Text)

	// We draw the text on the image.
	d := &font.Drawer{
		Dst:  m,
		Src:  image.NewUniform(p.Foreground),
		Face: face,
		Dot:  fixed.Point26_6{X: (fixed.I(p.Width) - textWidth) / 2, Y: fixed.I(500)},
	}
	d.DrawString(p.Text)
	return m, nil
}

// loadFont loads a TrueType font from the given path.
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

// loadImg loads an image given its path.
func loadImg(path string) (image.Image, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("could not open %s: %v", path, err)
	}
	defer f.Close()

	m, _, err := image.Decode(f)
	if err != nil {
		return nil, fmt.Errorf("could not decode %s: %v", path, err)
	}
	return m, nil
}

// WriteImg encodes an image and stores it in a file with the given path.
// fitFontSize finds the font size at which the given text fits 80% of the
// given width and the indentation required to center the text.
func fitFontSize(f *truetype.Font, width int, text string) (font.Face, fixed.Int26_6) {
	fixw := fixed.I(width)
	curWidth := 2 * fixw
	var face font.Face

	// We find which is the good font size for the width.
	for size := 100.0; curWidth > fixw; size-- {
		face = truetype.NewFace(f, &truetype.Options{
			Size:    size,
			Hinting: font.HintingNone,
			DPI:     72,
		})
		curWidth = (&font.Drawer{Face: face}).MeasureString(text)
	}

	return face, curWidth
}
