package main

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"strconv"
	"strings"

	"github.com/campoy/podcast-to-youtube/podcast"

	"golang.org/x/oauth2/google"

	youtube "google.golang.org/api/youtube/v3"
)

const (
	rssFeed    = "http://feeds.feedburner.com/GcpPodcast?format=xml"
	playlistID = "PLIivdWyY5sqJOTOszXDZh3XustjvTsrmQ"
)

func main() {
	eps, err := podcast.FetchFeed(rssFeed)
	if err != nil {
		log.Fatal(err)
	}

	last, err := fetchLastPublished()
	if err != nil {
		log.Fatal(err)
	}

	for i := len(eps) - 1; i >= 0; i-- {
		if eps[i].Number == last {
			eps = eps[i+1:]
			break
		}
	}

	for _, ep := range eps {
		fmt.Println(ep.Title)
	}
}

func fetchLastPublished() (int, error) {
	ctx := context.Background()

	data, err := ioutil.ReadFile("secret.json")
	if err != nil {
		return 0, fmt.Errorf("could not read service account file: %v", err)
	}

	cfg, err := google.JWTConfigFromJSON(data, youtube.YoutubeScope, youtube.YoutubeReadonlyScope)
	if err != nil {
		return 0, fmt.Errorf("could not create authenticated client: %v", err)
	}

	yt, err := youtube.New(cfg.Client(ctx))
	if err != nil {
		return 0, fmt.Errorf("could not create YouTube client: %v", err)
	}

	res, err := yt.PlaylistItems.List("snippet").PlaylistId(playlistID).MaxResults(1).Do()
	if err != nil {
		return 0, fmt.Errorf("could not fetch playlist: %v", err)
	}

	if len(res.Items) == 0 {
		return 0, fmt.Errorf("playlist is empty")
	}

	title := res.Items[0].Snippet.Title
	num, err := strconv.Atoi(title[strings.LastIndex(title, " ")+1:])
	if err != nil {
		return 0, fmt.Errorf("could not find number in title %q: %v", title, err)
	}
	return num, nil
}
