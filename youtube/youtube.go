package youtube

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"

	"golang.org/x/oauth2/google"

	"golang.org/x/oauth2"
	youtube "google.golang.org/api/youtube/v3"
)

type Client struct{ svc *youtube.Service }

// NewClient creates a new authenticated client given the path of an oauth2 secret service and a token.
func NewClient(secret, token string) (*Client, error) {
	data, err := ioutil.ReadFile(token)
	if err != nil {
		return nil, fmt.Errorf("could not open %s: %v", token, err)
	}
	var tok oauth2.Token
	if err := json.Unmarshal(data, &tok); err != nil {
		return nil, fmt.Errorf("could not parse token: %v", err)
	}
	data, err = ioutil.ReadFile(secret)
	if err != nil {
		return nil, fmt.Errorf("could not open %s: %v", secret, err)
	}
	cfg, err := google.ConfigFromJSON(data, youtube.YoutubeScope, youtube.YoutubeReadonlyScope, youtube.YoutubeUploadScope)
	if err != nil {
		return nil, fmt.Errorf("could not parse config: %v", err)
	}
	svc, err := youtube.New(cfg.Client(context.Background(), &tok))
	if err != nil {
		return nil, fmt.Errorf("could not create youtube client: %v", err)
	}
	return &Client{svc}, nil
}

// FetchLastPublished finds the number of the latest episode published in the playlist.
func (client *Client) FetchLastPublished(playlistID string) (*youtube.PlaylistItem, error) {
	res, err := client.svc.PlaylistItems.List("snippet").PlaylistId(playlistID).MaxResults(1).Do()
	if err != nil {
		return nil, fmt.Errorf("could not fetch playlist: %v", err)
	}

	if len(res.Items) == 0 {
		return nil, fmt.Errorf("playlist is empty")
	}
	return res.Items[0], nil
}

// Upload uploads the video in the given path to YouTube with the given details.
func (client *Client) Upload(title, desc string, tags []string, path string) error {
	f, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("could not open %v: %v", path, err)
	}
	defer f.Close()

	video := &youtube.Video{
		Snippet: &youtube.VideoSnippet{
			Title:       title,
			Description: desc,
			Tags:        tags,
		},
		Status: &youtube.VideoStatus{PrivacyStatus: "unlisted"},
	}
	call := client.svc.Videos.Insert("snippet,status", video)
	_, err = call.Media(f).Do()
	return err
}
