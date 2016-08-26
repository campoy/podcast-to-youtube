package main

import (
	"bytes"
	"context"
	"encoding/xml"
	"flag"
	"fmt"
	"net/http"
	"strings"
)

func main() {
	var (
		rssFeed   = flag.String("rss", "http://feeds.feedburner.com/GcpPodcast?format=xml", "url for the RSS feed")
		logo      = flag.String("logo", "logo.png", "path to the PNG logo image")
		titleTmpl = flag.String("title", "%s: GCPPodcast %d", "template used for the title")
	)
	flag.Parse()

	fmt.Print("episode number to publish: ")
	var number int
	fmt.Scanf("%d", &number)

	ep, err := fetchEpisode(*rssFeed, number)
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Printf("episode %d: %s\n", ep.Number, ep.Title)
	fmt.Print("publish? (Y/n): ")
	var answer string
	fmt.Scanln(&answer)
	if !(answer == "Y" || answer == "y" || answer == "") {
		return
	}

	ctx := context.Background()

	vid, err := createVideo(*logo, fmt.Sprintf("%d: %s", ep.Number, ep.Title), ep.MP3)
	if err != nil {
		fmt.Printf("could not create video: %v\n", err)
		return
	}

	title := fmt.Sprintf(*titleTmpl, ep.Title, ep.Number)
	desc := fmt.Sprintf("Original post: %s\n\n", ep.Link) + dropHTMLTags(ep.Desc)
	tags := append(ep.Tags, "gcppodcast", "podcast")

	if err := uploadToYouTube(ctx, title, desc, tags, vid); err != nil {
		fmt.Println(err)
		return
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

func fetchEpisode(rss string, number int) (*episode, error) {
	res, err := http.Get(rss)
	if err != nil {
		return nil, fmt.Errorf("could not get %s: %v", rss, err)
	}
	defer res.Body.Close()

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

	for _, i := range data.Channel[0].Item {
		if i.Number == number {
			return &episode{
				Title:  i.Title,
				Number: i.Number,
				Link:   i.Link,
				Desc:   i.Desc,
				MP3:    i.MP3.URL,
				Tags:   i.Category,
			}, nil
		}
	}

	return nil, fmt.Errorf("could not find episode %d", number)
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
