// Copyright 2016 Google Inc. All rights reserved.
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to writing, software distributed
// under the License is distributed on a "AS IS" BASIS, WITHOUT WARRANTIES OR
// CONDITIONS OF ANY KIND, either express or implied.
//
// See the License for the specific language governing permissions and
// limitations under the License.

package podcast2youtube

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"

	"google.golang.org/api/youtube/v3"
	"gopkg.in/cheggaaa/pb.v1"
)

// UploadToYouTube uploads the video in the given path to YouTube with the given
// details. This will prompt an offline authentication flow.
func UploadToYouTube(client *http.Client, title, desc string, tags []string, path string) error {
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
	defer func() { _ = f.Close() }()

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
