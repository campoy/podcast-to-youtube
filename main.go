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

// The podcast2youtube command uses ffmpeg to generate videos from any given
// podcast, by downloading the mp3 and adding a fix image with a given logo
// and text.
package main

import (
	"bytes"
	"context"
	"flag"
	"fmt"
	"image"
	"image/png"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/campoy/podcast2youtube/podcast2youtube"
)

var (
	rssFeed   = flag.String("rss", "http://feeds.feedburner.com/GcpPodcast?format=xml", "url for the RSS feed")
	logo      = flag.String("logo", "logo.png", "path to the PNG logo image")
	titleTmpl = flag.String("title", "%s: GCPPodcast %d", "template used for the title")
	fgHex     = flag.String("fg", "009688", "hex encoded color for the video text")
	bgHex     = flag.String("bg", "ffffff", "hex encoded color for the video background")
	width     = flag.Int("w", 1200, "width of the generated video in pixels")
	height    = flag.Int("h", 800, "height of the generated video in pixels")
)

func main() {
	flag.Parse()

	fmt.Print("episode number to publish: ")
	var number int
	fmt.Scanf("%d", &number)

	ep, err := fetchEpisode(*rssFeed, number)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	fmt.Printf("episode %d: %s\n", ep.Number, ep.Title)
	fmt.Print("publish? (Y/n): ")
	var answer string
	fmt.Scanln(&answer)
	if !(answer == "Y" || answer == "y" || answer == "") {
		return
	}

	ctx := context.Background()
	if err := buildAndUpload(ctx, *logo, ep); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func buildAndUpload(ctx context.Context, logo string, ep *episode) error {
	tmpDir, err := ioutil.TempDir("", "")
	if err != nil {
		return fmt.Errorf("could not create temp directory: %v", err)
	}
	defer func() {
		if err := os.RemoveAll(tmpDir); err != nil {
			log.Printf("could not remove %s: %v", tmpDir, err)
		}
	}()

	l, err := loadPNG(logo)
	if err != nil {
		return fmt.Errorf("could not load logo %s: %v", logo, err)
	}

	text := fmt.Sprintf("%d: %s", ep.Number, ep.Title)
	m, err := podcast2youtube.CreateImage(l, text, *fgHex, *bgHex, *width, *height)
	if err != nil {
		return fmt.Errorf("could not create image: %v", err)
	}

	slide := filepath.Join(tmpDir, "slide.png")
	if err := writePNG(slide, m); err != nil {
		return err
	}

	vid := filepath.Join(tmpDir, "vid.mp4")
	if err := podcast2youtube.CreateVideo(slide, ep.MP3, vid); err != nil {
		return fmt.Errorf("could not create video: %v\n", err)
	}

	title := fmt.Sprintf(*titleTmpl, ep.Title, ep.Number)
	desc := fmt.Sprintf("Original post: %s\n\n", ep.Link) + dropHTMLTags(ep.Desc)
	tags := append(ep.Tags, "gcppodcast", "podcast")

	if err := podcast2youtube.UploadToYouTube(ctx, title, desc, tags, vid); err != nil {
		return fmt.Errorf("could not upload to YouTube: %v", err)
	}
	return nil
}

func dropHTMLTags(s string) string {
	w := bytes.Buffer{}
	inTag := false
	for _, r := range s {
		switch {
		case !inTag && r == '<':
			inTag = true
		case inTag && r == '>':
			inTag = false
			continue
		}
		if !inTag {
			fmt.Fprintf(&w, "%c", r)
		}
	}
	return strings.Replace(w.String(), "\n", " ", -1)
}

func loadPNG(path string) (m image.Image, err error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("could not open %s: %v", path, err)
	}
	defer func() {
		if cerr := f.Close(); cerr != nil && err == nil {
			err = fmt.Errorf("could not close %s: %v", path, err)
		}
	}()

	m, err = png.Decode(f)
	if err != nil {
		return nil, fmt.Errorf("could not decode %s: %v", path, err)
	}
	return m, nil
}

func writePNG(path string, m image.Image) error {
	f, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("could not create %s: %v", path, err)
	}
	defer func() {
		if cerr := f.Close(); cerr != nil && err == nil {
			err = fmt.Errorf("could not close %s: %v", path, err)
		}
	}()

	if err := png.Encode(f, m); err != nil {
		return fmt.Errorf("could not encode to %s: %v", path, err)
	}
	return nil
}
