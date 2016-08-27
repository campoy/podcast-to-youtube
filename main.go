// The podcast2youtube command uses ffmpeg to generate videos from any given
// podcast, by downloading the mp3 and adding a fix image with a given logo
// and text.
package main

import (
	"bytes"
	"context"
	"encoding/xml"
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"golang.org/x/oauth2/google"
	"google.golang.org/api/youtube/v3"

	"github.com/campoy/podcast2youtube/podcast2youtube"
)

var (
	rssFeed   = flag.String("rss", "http://feeds.feedburner.com/GcpPodcast?format=xml", "url for the RSS feed")
	logo      = flag.String("logo", "resources/logo.png", "path to the PNG logo image")
	font      = flag.String("font", "resources/Roboto-Light.ttf", "font to be used in the video")
	titleTmpl = flag.String("title", "%s: GCPPodcast %d", "template used for the title")
	fgHex     = flag.String("fg", "ffffff", "hex encoded color for the video text")
	bgHex     = flag.String("bg", "009688", "hex encoded color for the video background")
	width     = flag.Int("w", 1200, "width of the generated video in pixels")
	height    = flag.Int("h", 800, "height of the generated video in pixels")
)

func main() {
	flag.Parse()

	eps, err := fetchFeed(*rssFeed)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	fmt.Print("episode number to publish (try 1, or 2-10): ")
	var answer string
	fmt.Scanln(&answer)
	from, to, err := parseRange(answer)
	if err != nil {
		fmt.Printf("%s is an invalid range\n", answer)
		return
	}

	var selected []episode
	for _, e := range eps {
		if from <= e.Number && e.Number <= to {
			selected = append(selected, e)
			fmt.Printf("episode %d: %s\n", e.Number, e.Title)
		}
	}

	fmt.Print("publish? (Y/n): ")
	answer = ""
	fmt.Scanln(&answer)
	if !(answer == "Y" || answer == "y" || answer == "") {
		return
	}

	if err := buildAndUpload(selected); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

type episode struct {
	Title  string
	Number int
	Link   string
	Desc   string
	MP3    string
	Tags   []string
}

func fetchFeed(rss string) ([]episode, error) {
	res, err := http.Get(rss)
	if err != nil {
		return nil, fmt.Errorf("could not get %s: %v", rss, err)
	}
	defer func() { _ = res.Body.Close() }()

	var data struct {
		XMLName xml.Name `xml:"rss"`
		Channel []struct {
			Item []struct {
				Title  string `xml:"title"`
				Number int    `xml:"order"`
				Link   string `xml:"guid"`
				Desc   string `xml:"summary"`
				MP3    struct {
					URL string `xml:"url,attr"`
				} `xml:"enclosure"`
				Category []string `xml:"category"`
			} `xml:"item"`
		} `xml:"channel"`
	}

	if err := xml.NewDecoder(res.Body).Decode(&data); err != nil {
		return nil, fmt.Errorf("could not decode feed: %v", err)
	}

	var eps []episode
	for _, i := range data.Channel[0].Item {
		eps = append(eps, episode{
			Title:  i.Title,
			Number: i.Number,
			Link:   i.Link,
			Desc:   i.Desc,
			MP3:    i.MP3.URL,
			Tags:   i.Category,
		})
	}
	return eps, nil
}

func parseRange(s string) (int, int, error) {
	switch ps := strings.Split(s, "-"); len(ps) {
	case 1:
		n, err := strconv.Atoi(ps[0])
		return n, n, err
	case 2:
		from, err := strconv.Atoi(ps[0])
		if err != nil {
			return 0, 0, err
		}
		to, err := strconv.Atoi(ps[1])
		return from, to, err
	default:
		return 0, 0, errors.New("only formats supported are n or m-n")
	}
}

func buildAndUpload(eps []episode) error {
	client, err := authedClient()
	if err != nil {
		return fmt.Errorf("could not authenticate: %v\n", err)
	}

	for _, ep := range eps {
		if err := buildAndUploadOne(client, ep); err != nil {
			return fmt.Errorf("episode %d: %v", ep.Number, err)
		}
	}
	return nil
}

func buildAndUploadOne(client *http.Client, ep episode) error {
	tmpDir, err := ioutil.TempDir("", "")
	if err != nil {
		return fmt.Errorf("could not create temp directory: %v", err)
	}
	defer func() {
		if err := os.RemoveAll(tmpDir); err != nil {
			log.Printf("could not remove %s: %v", tmpDir, err)
		}
	}()

	img := podcast2youtube.Image{
		Logo:       *logo,
		Text:       fmt.Sprintf("%d: %s", ep.Number, ep.Title),
		Font:       *font,
		Foreground: *fgHex,
		Background: *bgHex,
		Width:      *width,
		Height:     *height,
	}

	// create the image
	slide := filepath.Join(tmpDir, "slide.png")
	if err := img.CreateIn(slide); err != nil {
		return fmt.Errorf("could not create image: %v", err)
	}

	// create the video
	vid := filepath.Join(tmpDir, "vid.mp4")
	if err := podcast2youtube.CreateVideo(slide, ep.MP3, vid); err != nil {
		return fmt.Errorf("could not create video: %v\n", err)
	}

	title := fmt.Sprintf(*titleTmpl, ep.Title, ep.Number)
	desc := fmt.Sprintf("Original post: %s\n\n", ep.Link) + dropHTMLTags(ep.Desc)
	tags := append(ep.Tags, "gcppodcast", "podcast")

	if err := podcast2youtube.UploadToYouTube(client, title, desc, tags, vid); err != nil {
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

// authedClient performs an offline OAuth flow.
func authedClient() (*http.Client, error) {
	const path = "client_secrets.json"
	b, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("could not open %s: %v", path, err)
	}
	cfg, err := google.ConfigFromJSON(b, youtube.YoutubeUploadScope)
	if err != nil {
		return nil, fmt.Errorf("could not parse config: %v", err)
	}

	url := cfg.AuthCodeURL("")
	fmt.Printf("Go here: \n\t%s\n", url)
	fmt.Printf("Then enter the code: ")
	var code string
	fmt.Scanln(&code)
	ctx := context.Background()
	tok, err := cfg.Exchange(ctx, code)
	if err != nil {
		return nil, err
	}
	return cfg.Client(ctx, tok), nil
}
