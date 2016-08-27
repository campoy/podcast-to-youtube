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
	"os"
	"os/exec"
)

// CreateVideo creates a video that plays the given audio with the given image
// as fixed background in the requested path.
func CreateVideo(img, mp3, vid string) error {
	const (
		slidePath = "slide.png"
		mp3Path   = "audio.mp3"
		vidPath   = "vid.mp4"
	)

	// ffmpeg -y -i slide.png -i audio.mp3 -pix_fmt yuv420p -c:a aac -c:v libx264 -crf 18 out.mp4
	cmd := exec.Command("ffmpeg", "-y", "-loop", "1", "-i", img, "-i", mp3, "-shortest",
		"-c:v", "libx264", "-pix_fmt", "yuv420p", "-c:a", "aac", "-crf", "18",
		vid)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
