package main

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"net/http"
	"os"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"

	"google.golang.org/api/youtube/v3"
)

func uploadToYouTube(ctx context.Context, number int, title, link, desc, path string) error {
	client, err := authedClient(ctx)
	if err != nil {
		return fmt.Errorf("could not authenticate: %v", err)
	}
	service, err := youtube.New(client)
	if err != nil {
		return fmt.Errorf("could not create YouTube client: %v", err)
	}

	upload := &youtube.Video{
		Snippet: &youtube.VideoSnippet{
			Title:       fmt.Sprintf("GCPPodcast #%d: %s", number, title),
			Description: dropHTMLTags(desc),
		},
		Status: &youtube.VideoStatus{PrivacyStatus: "unlisted"},
	}

	f, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("could not open %v: %v", path, err)
	}
	defer f.Close()

	fmt.Println("uploading video to YouTube")

	r, err := progressBarReader(f)
	if err != nil {
		log.Printf("could not create progress bar: %v", err)
		r = f
	}

	call := service.Videos.Insert("snippet,status", upload)
	if _, err := call.Media(r).Do(); err != nil {
		return fmt.Errorf("could not upload: %v", err)
	}

	return nil
}

// authedClient performs an OAuth flow.
func authedClient(ctx context.Context) (*http.Client, error) {
	url := oauthConfig.AuthCodeURL("")
	fmt.Printf("Go here: \n\t%s\n", url)
	fmt.Printf("Then enter the code: ")
	var code string
	fmt.Scanln(&code)
	tok, err := oauthConfig.Exchange(context.Background(), code)
	if err != nil {
		return nil, err
	}

	return oauthConfig.Client(ctx, tok), nil
}

var oauthConfig = &oauth2.Config{
	ClientID:     "689203460032-7oobm6aedva96ni27argap7l3gd8np6b.apps.googleusercontent.com",
	ClientSecret: "aSH7VNMy5l_cR4OhUtx1RXvb",
	Scopes:       []string{youtube.YoutubeUploadScope},
	Endpoint:     google.Endpoint,
	RedirectURL:  "oob",
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
	return w.String()
}
