package main

import (
	"context"
	"encoding/xml"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"

	"gopkg.in/cheggaaa/pb.v1"
)

const (
	bucketName = "podcast-to-youtube"
	rssFeed    = "http://feeds.feedburner.com/GcpPodcast?format=xml"
)

func main() {
	fmt.Print("episode number to publish: ")
	var number int
	fmt.Scanf("%d", &number)

	ep, err := fetchEpisode(number)
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Printf("episode %d: %s\n", ep.Number, ep.Title)

	if !question("publish") {
		return
	}

	ctx := context.Background()

	var vid string
	if videoExists(ctx, ep.Number) && question("video already exists: want to reuse it") {
		vid, err = downloadVideo(ctx, ep.Number)
		if err != nil {
			fmt.Printf("could not download video: %v\n", err)
			return
		}
	} else {
		vid, err = createVideo(ep.Number, ep.Title, ep.MP3)
		if err != nil {
			fmt.Printf("could not create video: %v\n", err)
			return
		}
	}

	if err := uploadVideo(ctx, ep.Number, vid); err != nil {
		log.Fatal(err)
	}

	if err := uploadToYouTube(ctx, ep.Number, ep.Title, ep.Link, ep.Desc, vid); err != nil {
		fmt.Println(err)
		return
	}
}

func question(text string) bool {
	fmt.Printf("%s? (Y/n): ", text)
	var answer string
	fmt.Scanln(&answer)
	return answer == "Y" || answer == "y" || answer == ""
}

type episode struct {
	Title  string
	Number int
	Link   string
	Desc   string
	MP3    string
}

func fetchEpisode(number int) (*episode, error) {
	res, err := http.Get(rssFeed)
	if err != nil {
		return nil, fmt.Errorf("could not get %s: %v", rssFeed, err)
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
			}, nil
		}
	}

	return nil, fmt.Errorf("could not find episode %d", number)
}

func progressBarReader(f *os.File) (io.Reader, error) {
	fi, err := f.Stat()
	if err != nil {
		return nil, fmt.Errorf("could not stat: %v", err)
	}
	bar := pb.StartNew(int(fi.Size())).SetUnits(pb.U_BYTES)
	bar.ShowSpeed = true
	bar.ShowTimeLeft = true
	bar.Start()
	return bar.NewProxyReader(f), nil
}
