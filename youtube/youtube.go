package youtube

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"time"

	"golang.org/x/oauth2/google"

	"golang.org/x/oauth2"
	youtube "google.golang.org/api/youtube/v3"
)

// Client provides methods to access the YouTube API.
type Client struct {
	svc *youtube.Service
	log func(string, ...interface{})
}

// NewClient creates a new authenticated client given the path of an oauth2 secret service and a token.
func NewClient(secret, token string, log func(string, ...interface{})) (*Client, error) {
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
	if log == nil {
		log = func(string, ...interface{}) {}
	}
	return &Client{svc, log}, nil
}

// FetchLastPublished finds the number of the latest episode published in the playlist.
func (c *Client) FetchLastPublished(playlist string) (*youtube.PlaylistItem, error) {
	res, err := c.svc.PlaylistItems.List("snippet").PlaylistId(playlist).MaxResults(1).Do()
	if err != nil {
		return nil, fmt.Errorf("could not fetch playlist: %v", err)
	}

	if len(res.Items) == 0 {
		return nil, fmt.Errorf("playlist is empty")
	}
	return res.Items[0], nil
}

// Upload uploads the video in the given path to YouTube with the given details.
func (c *Client) Upload(title, desc string, tags []string, path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", fmt.Errorf("could not open %v: %v", path, err)
	}
	defer f.Close()

	v := &youtube.Video{
		Snippet: &youtube.VideoSnippet{
			Title:       title,
			Description: desc,
			Tags:        tags,
		},
		Status: &youtube.VideoStatus{PrivacyStatus: "public"},
	}
	v, err = c.svc.Videos.Insert("snippet,status", v).Media(f).Do()
	if err != nil {
		return "", fmt.Errorf("could not insert video: %v", err)
	}
	return v.Id, nil
}

// AddToPlaylist adds the given video id to a plyalist.
func (c *Client) AddToPlaylist(playlist, video string) error {
	call := c.svc.PlaylistItems.Insert("snippet", &youtube.PlaylistItem{
		Snippet: &youtube.PlaylistItemSnippet{
			PlaylistId: playlist,
			ResourceId: &youtube.ResourceId{
				VideoId: video,
				Kind:    "youtube#video",
			},
		},
	})
	_, err := call.Do()
	return err
}

// Status returns the current status of a YouTube video.
func (c *Client) Status(video string) (string, error) {
	res, err := c.svc.Videos.List("status").Id(video).Do()
	if err != nil {
		return "", err
	}
	if len(res.Items) != 1 {
		return "", fmt.Errorf("expected one item in response; got %d", len(res.Items))
	}
	return res.Items[0].Status.UploadStatus, nil
}

// WaitUntilProcessed blocks until the video is processed or the given context is canceled.
func (c *Client) WaitUntilProcessed(ctx context.Context, video string) error {
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			s, err := c.Status(video)
			if err != nil {
				return fmt.Errorf("could not check status")
			}
			c.log("status of video is %q", s)
			if s == "processed" {
				return nil
			}
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}
