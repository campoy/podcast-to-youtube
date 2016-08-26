package main

import (
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"

	"golang.org/x/oauth2/google"
	"google.golang.org/api/youtube/v3"
	"gopkg.in/cheggaaa/pb.v1"
)

func uploadToYouTube(ctx context.Context, title, desc string, tags []string, path string) error {
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
			Title:       title,
			Description: desc,
			Tags:        tags,
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

// authedClient performs an offline OAuth flow.
func authedClient(ctx context.Context) (*http.Client, error) {
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
	tok, err := cfg.Exchange(context.Background(), code)
	if err != nil {
		return nil, err
	}
	return cfg.Client(ctx, tok), nil
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
