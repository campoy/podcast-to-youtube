package main

import (
	"fmt"
	"net/http"
	"os"

	"google.golang.org/api/youtube/v3"
)

func uploadToYouTube(number, title, path string) error {
	service, err := youtube.New(http.DefaultClient)
	if err != nil {
		return fmt.Errorf("could not create YouTube client: %v", err)
	}

	upload := &youtube.Video{
		Snippet: &youtube.VideoSnippet{
			Title:       fmt.Sprintf("GCPPodcast #%s: %s", number, title),
			Description: "some description and cool stuff",
		},
		Status: &youtube.VideoStatus{PrivacyStatus: "unlisted"},
	}

	f, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("could not open %v: %v", path, err)
	}
	defer f.Close()

	call := service.Videos.Insert("snippet,status", upload)
	if _, err := call.Media(f).Do(); err != nil {
		return fmt.Errorf("could not upload: %v", err)
	}
	return nil
}

var clientID = `
{
   "web" : {
      "auth_uri" : "https://accounts.google.com/o/oauth2/auth",
      "client_id" : "689203460032-0tupsmifoou6put7plg0ka6lgefjaprt.apps.googleusercontent.com",
      "token_uri" : "https://accounts.google.com/o/oauth2/token",
      "client_secret" : "hO64wKfWicTg4VEH4wbHbVes",
      "project_id" : "podcast-to-youtube",
      "auth_provider_x509_cert_url" : "https://www.googleapis.com/oauth2/v1/certs"
   }
}
`
