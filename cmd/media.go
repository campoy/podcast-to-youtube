package main

import (
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"

	"github.com/golang/freetype/truetype"
	"golang.org/x/image/font"
	"golang.org/x/image/math/fixed"
)

func createVideo(number int, title, mp3 string) (string, error) {
	const (
		slidePath = "slide.png"
		mp3Path   = "audio.mp3"
		vidPath   = "vid.mp4"
	)

	// create a new file, will truncate if existing.
	f, err := os.Create(slidePath)
	if err != nil {
		return "", fmt.Errorf("could not create slide.png: %v", err)
	}
	defer f.Close()
	defer os.Remove(slidePath)

	// create the background image for the video and writing to slide.png.
	m, err := createImage(number, title)
	if err != nil {
		return "", err
	}
	if err := png.Encode(f, m); err != nil {
		return "", fmt.Errorf("could not encode image: %v", err)
	}

	// download the mp3 and save to audio.mp3
	res, err := http.Get(mp3)
	if err != nil {
		return "", fmt.Errorf("could not download audio %s: %v", mp3, err)
	}
	defer res.Body.Close()

	f, err = os.Create(mp3Path)
	if err != nil {
		return "", fmt.Errorf("could not create audio.mp3: %v", err)
	}
	defer f.Close()
	defer os.Remove(mp3Path)

	if _, err := io.Copy(f, res.Body); err != nil {
		return "", fmt.Errorf("could not write to audio.mp3: %v", err)
	}

	// ffmpeg -y -i slide.png -i audio.mp3 -pix_fmt yuv420p -c:a aac -c:v libx264 -crf 18 out.mp4
	cmd := exec.Command("ffmpeg", "-y", "-loop", "1", "-i", slidePath, "-i", mp3Path, "-shortest",
		"-c:v", "libx264", "-pix_fmt", "yuv420p", "-c:a", "aac", "-crf", "18",
		vidPath)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("ffmpeg failed")
	}
	return vidPath, nil
}

func createImage(number int, title string) (image.Image, error) {
	text := fmt.Sprintf("%v: %s", number, title)

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
