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

package image

import (
	"image"
	"image/color"
	"image/png"
	"log"
	"os"
	"os/exec"
	"testing"
)

var goldenImage = func() image.Image {
	m, err := loadImg("golden.png")
	if err != nil {
		log.Fatal(err)
	}
	return m
}()

func TestGenerate(t *testing.T) {
	p := Params{
		Logo:       "../resources/logo.png",
		Text:       "42: this is a test",
		Font:       "../resources/Roboto-Light.ttf",
		Foreground: color.White,
		Background: color.RGBA{50, 100, 50, 255},
		Width:      1280,
		Height:     720,
	}
	m, err := Generate(p)
	if err != nil {
		t.Fatalf("could not generate: %v", err)
	}
	checkImagesEq(t, goldenImage, m)
}

func checkImagesEq(t *testing.T, a, b image.Image) {
	if ac, bc := a.ColorModel(), b.ColorModel(); ac != bc {
		t.Errorf("different color models: wanted %v got %v", ac, bc)
		return
	}
	if ba, bb := a.Bounds(), b.Bounds(); ba != bb {
		t.Errorf("different image sizes: wanted %v got %v", ba, bb)
		return
	}
	diff := false
	for i := a.Bounds().Min.X; i <= a.Bounds().Max.X; i++ {
		for j := a.Bounds().Min.Y; j <= a.Bounds().Max.Y; j++ {
			if a.At(i, j) != b.At(i, j) {
				b.(*image.RGBA).Set(i, j, color.RGBA{255, 0, 0, 255})
				diff = true
			}
		}
	}
	if !diff {
		return
	}

	t.Errorf("different color at least one pixel")
	f, err := os.Create("diff.png")
	if err != nil {
		t.Fatalf("could not create diff image: %v", err)
	}
	defer f.Close()
	if err := png.Encode(f, b); err != nil {
		t.Fatalf("could not encode diff image: %v", err)
	}
	t.Errorf("see differences as red pixels on diff.png")
	exec.Command("open", "diff.png").Run()
}
