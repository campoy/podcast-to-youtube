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

package podcast

import (
	"encoding/xml"
	"fmt"
	"net/http"
	"sort"
)

// An Episode contains all the information available for a podcast
// in a RSS stream.
type Episode struct {
	Title  string
	Number int
	Link   string
	Desc   string
	MP3    string
	Tags   []string
}

// FetchFeed fetches a list of episodes for a podcast given its RSS feed URL.
func FetchFeed(rss string) ([]Episode, error) {
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

	var eps []Episode
	for _, i := range data.Channel[0].Item {
		eps = append(eps, Episode{
			Title:  i.Title,
			Number: i.Number,
			Link:   i.Link,
			Desc:   i.Desc,
			MP3:    i.MP3.URL,
			Tags:   i.Category,
		})
	}

	sort.Slice(eps, func(i, j int) bool { return eps[i].Number < eps[j].Number })
	return eps, nil
}
