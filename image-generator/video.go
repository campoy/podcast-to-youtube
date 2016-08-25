package main

import (
	"fmt"
	"os/exec"
)

func createVideo(img, mp3 string) (string, error) {
	out := "video.mp4"
	// ffmpeg -loop 1 -i slide.png -i audio.mp3 -c:v libx264 -pix_fmt yuv420p video.mp4
	cmd := exec.Command("ffmpeg", "-i", img, "-i", mp3,
		"-c:v", "libx264", "-pix_fmt", "yuv420p",
		out)

	if b, err := cmd.CombinedOutput(); err != nil {
		return "", fmt.Errorf("ffmpeg failed: %s", b)
	}
	return out, nil
}
