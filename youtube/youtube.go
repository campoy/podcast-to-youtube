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
func (client *Client) Upload(title, desc string, tags []string, path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", fmt.Errorf("could not open %v: %v", path, err)
	}
	defer f.Close()

	video := &youtube.Video{
		Snippet: &youtube.VideoSnippet{
			Title:       title,
			Description: desc,
			Tags:        tags,
		},
		Status: &youtube.VideoStatus{PrivacyStatus: "public"},
	}
	call := client.svc.Videos.Insert("snippet,status", video)
	video, err = call.Media(f).Do()
	if err != nil {
		return "", fmt.Errorf("could not insert video: %v", err)
	}
	return video.Id, nil
}

// AddToPlaylist adds the given video id to a plyalist.
func (client *Client) AddToPlaylist(playlistID, videoID string) error {
	call := client.svc.PlaylistItems.Insert("snippet", &youtube.PlaylistItem{
		Snippet: &youtube.PlaylistItemSnippet{
			PlaylistId: playlistID,
			ResourceId: &youtube.ResourceId{
				VideoId: videoID,
				Kind:    "youtube#video",
			},
		},
	})
	_, err := call.Do()
	return err
}

// IsProcessed returns whether a video has been successfully processed.
func (client *Client) Status(videoID string) (*youtube.VideoStatus, error) {
	res, err := client.svc.Videos.List("status").Id(videoID).Do()
	if err != nil {
		return nil, err
	}
	if len(res.Items) != 1 {
		return nil, fmt.Errorf("expected one item in response; got %d", len(res.Items))
	}
	return res.Items[0].Status, nil
}

// WaitForProcessed blocks until the video is processed.
func (client *Client) WaitForProcessed(videoID string, timeout time.Duration, log func(string, ...interface{})) error {
	done := time.After(timeout)
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			s, err := client.Status(videoID)
			if err != nil {
				return fmt.Errorf("could not check status")
			}
			if s.UploadStatus == "processed" {
				return nil
			}
		case <-done:
			return fmt.Errorf("video not processed after %v", timeout)
		}
	}
}
